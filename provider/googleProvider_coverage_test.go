package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type requestRecord struct {
	Method string
	URL    string
	Body   []byte
}

type responseStep struct {
	body    string
	validate func(*testing.T, requestRecord)
}

type scriptedTransport struct {
	t          *testing.T
	responses  []responseStep
	mu         sync.Mutex
	requests   []requestRecord
}

func (s *scriptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}

	record := requestRecord{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   body,
	}

	s.mu.Lock()
	idx := len(s.requests)
	s.requests = append(s.requests, record)
	s.mu.Unlock()

	if idx >= len(s.responses) {
		s.t.Fatalf("unexpected request %d: %s %s", idx, record.Method, record.URL)
	}

	step := s.responses[idx]
	if step.validate != nil {
		step.validate(s.t, record)
	}

	respBody := step.body
	if respBody == "" {
		respBody = "{}"
	}

	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(respBody)),
		Request:    req,
	}, nil
}

func newTestGoogleProvider(t *testing.T, responses ...responseStep) (*GoogleProvider, *scriptedTransport) {
	t.Helper()

	transport := &scriptedTransport{
		t:         t,
		responses: responses,
	}
	client := &http.Client{Transport: transport}
	svc, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	gp := &GoogleProvider{
		sheetsService:  svc,
		client:         client,
		sheetName:      "notes",
		sheetID:        "sheet-id",
		noteIds:        map[int]int{},
		notesUpdatedAt: map[int]int64{},
		idsMux:         &sync.RWMutex{},
		updAtMux:       &sync.RWMutex{},
		updateQueue:    make(chan *model.Note, 10),
		deleteQueue:    make(chan int, 10),
		ctx:            context.Background(),
		logger:         logger,
		observer:       &observer.ObserverImpl{},
	}
	return gp, transport
}

func TestGoogleProvider_CacheHelpers(t *testing.T) {
	t.Parallel()

	gp := &GoogleProvider{
		noteIds:        map[int]int{},
		notesUpdatedAt: map[int]int64{},
		idsMux:         &sync.RWMutex{},
		updAtMux:       &sync.RWMutex{},
	}

	gp.CacheIDSet(10, 3, false)
	idIdx, ok := gp.CacheIDGet(10)
	require.True(t, ok)
	assert.Equal(t, 3, idIdx)

	gp.CacheIDUnset(10)
	_, ok = gp.CacheIDGet(10)
	assert.False(t, ok)

	gp.CacheUpdAtSet(10, 123, false)
	updAt, ok := gp.CacheUpdAtGet(10)
	require.True(t, ok)
	assert.Equal(t, int64(123), updAt)

	gp.CacheUpdAtUnset(10)
	_, ok = gp.CacheUpdAtGet(10)
	assert.False(t, ok)
}

func TestGoogleProvider_InitAndConstructorErrors(t *testing.T) {
	t.Parallel()

	logger := logrus.New()
	logger.SetOutput(io.Discard)
	obs := &observer.ObserverImpl{}

	gp, err := NewGoogleProvider("", "", "", logger, obs)
	require.Error(t, err)
	assert.Nil(t, gp)

	gp = &GoogleProvider{
		sheetName:    "notes",
		sheetID:      "sheet-id",
		credFilePath: filepath.Join(t.TempDir(), "missing.json"),
	}
	err = gp.Init()
	require.Error(t, err)
}

func TestGoogleProvider_GetNoteIDs_UsesCacheWhenAllowed(t *testing.T) {
	t.Parallel()

	gp := &GoogleProvider{
		noteIds: map[int]int{1: 0},
	}

	ids, err := gp.GetNoteIDs(false)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, 0, ids[1])
}

func TestGoogleProvider_GetNote_NotFound(t *testing.T) {
	t.Parallel()

	gp := &GoogleProvider{
		noteIds: map[int]int{1: 0},
		idsMux:  &sync.RWMutex{},
		updAtMux: &sync.RWMutex{},
	}

	_, err := gp.GetNote(2)
	require.Error(t, err)
	assert.Equal(t, common.ERR_NOTE_NOT_FOUND, err.Error())
}

func TestGoogleProvider_GetNotes_FilterAndParse(t *testing.T) {
	gp, _ := newTestGoogleProvider(t, responseStep{
		body: `{"values":[["1","Alpha","Body A","false","false","","100","200"],["2","Beta","Body B","true","true","beta-key","300","400"]]}`,
	})

	notes, err := gp.GetNotes(2)
	require.NoError(t, err)
	require.Len(t, notes, 1)

	note := notes[0]
	assert.Equal(t, 2, note.ID)
	assert.Equal(t, "Beta", note.Title)
	assert.Equal(t, "Body B", note.Content)
	assert.True(t, note.Hidden)
	assert.True(t, note.Encrypted)
	assert.Equal(t, "beta-key", note.EncKeyName)
	assert.Equal(t, int64(300), note.CreatedAt)
	assert.Equal(t, int64(400), note.UpdatedAt)
}

func TestGoogleProvider_PutGetDeleteRoundTrip(t *testing.T) {
	gp, transport := newTestGoogleProvider(t,
		responseStep{
			body: `{"values":[["1"],["2"]]}`,
		},
		responseStep{
			body: `{"values":[["100"],["200"]]}`,
		},
		responseStep{
			validate: func(t *testing.T, rec requestRecord) {
				var payload struct {
					Values [][]interface{} `json:"values"`
				}
				require.NoError(t, json.Unmarshal(rec.Body, &payload))
				require.Len(t, payload.Values, 1)
				row := payload.Values[0]
				require.Len(t, row, 8)
				assert.EqualValues(t, 7, row[0])
				assert.Equal(t, "Alpha", row[1])
				assert.Equal(t, "Body", row[2])
				assert.Equal(t, true, row[3])
				assert.Equal(t, false, row[4])
				assert.Equal(t, "alpha-key", row[5])
				assert.EqualValues(t, 111, row[6])
				assert.EqualValues(t, 222, row[7])
			},
		},
		responseStep{
			body: `{"values":[["7","Alpha","Body","true","false","alpha-key","111","222"]]}`,
		},
		responseStep{
			validate: func(t *testing.T, rec requestRecord) {
				var payload struct {
					Values [][]interface{} `json:"values"`
				}
				require.NoError(t, json.Unmarshal(rec.Body, &payload))
				require.Len(t, payload.Values, 1)
				row := payload.Values[0]
				require.Len(t, row, 8)
				assert.EqualValues(t, 7, row[0])
				assert.Equal(t, "Alpha", row[1])
				assert.Equal(t, "Body updated", row[2])
				assert.Equal(t, true, row[3])
				assert.Equal(t, false, row[4])
				assert.Equal(t, "alpha-key", row[5])
				assert.EqualValues(t, 111, row[6])
				assert.EqualValues(t, 333, row[7])
			},
		},
		responseStep{},
	)

	newNote := &model.Note{
		ID:         7,
		Title:      "Alpha",
		Content:    "Body",
		Hidden:     true,
		Encrypted:  false,
		EncKeyName: "alpha-key",
		CreatedAt:  111,
		UpdatedAt:  222,
	}

	require.NoError(t, gp.PutNote(newNote))
	cacheIdx, ok := gp.CacheIDGet(newNote.ID)
	require.True(t, ok)
	assert.Equal(t, 2, cacheIdx)

	remote, err := gp.GetNote(newNote.ID)
	require.NoError(t, err)
	assert.Equal(t, 7, remote.ID)
	assert.Equal(t, "Alpha", remote.Title)
	assert.Equal(t, "Body", remote.Content)
	assert.True(t, remote.Hidden)
	assert.False(t, remote.Encrypted)
	assert.Equal(t, int64(111), remote.CreatedAt)
	assert.Equal(t, int64(222), remote.UpdatedAt)

	newNote.Content = "Body updated"
	newNote.UpdatedAt = 333
	require.NoError(t, gp.PutNote(newNote))

	require.NoError(t, gp.DeleteNote(newNote.ID))
	_, ok = gp.CacheIDGet(newNote.ID)
	assert.False(t, ok)

	require.Len(t, transport.requests, 6)
	assert.Contains(t, transport.requests[2].URL, "/values/")
	assert.Contains(t, transport.requests[3].URL, "/values/")
}

func TestGoogleProvider_SyncNotes(t *testing.T) {
	gp, _ := newTestGoogleProvider(t,
		responseStep{
			body: `{"values":[["1"],["2"]]}`,
		},
		responseStep{
			body: `{"values":[["100"],["200"]]}`,
		},
		responseStep{
			body: `{"values":[["1","Remote One","Remote Body","false","false","","100","100"]]}`,
		},
		responseStep{
			body: `{}`,
		},
		responseStep{
			body: `{"values":[["2","Remote Two","Other Body","true","false","","200","200"]]}`,
		},
	)

	titlesCh := make(chan []string, 1)
	gp.observer.AddListener(observer.EVENT_UPDATE_NOTE_TITLES, observer.Listener{
		OnNotify: func(data interface{}, args ...interface{}) {
			titles, ok := data.([]string)
			if ok {
				titlesCh <- titles
			}
		},
	})

	dbNotes := []model.Note{
		{
			ID:         1,
			Title:      "Local One",
			Content:    "Local One Body",
			CreatedAt:  10,
			UpdatedAt:  50,
			EncKeyName: "local-key",
		},
		{
			ID:         3,
			Title:      "Local Three",
			Content:    "Local Three Body",
			CreatedAt:  30,
			UpdatedAt:  300,
			EncKeyName: "local-key",
		},
	}

	downloaded, err := gp.SyncNotes(context.Background(), dbNotes)
	require.NoError(t, err)
	require.Len(t, downloaded, 1)
	assert.Equal(t, 2, downloaded[0].ID)
	assert.Equal(t, "Remote Two", downloaded[0].Title)

	select {
	case titles := <-titlesCh:
		assert.Equal(t, []string{"Remote One", "Local Three"}, titles)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for note titles update")
	}
}
