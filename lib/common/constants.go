package common

import (
	"path/filepath"
	"time"
)

// EncryptionKeyAction enum to describe the encryption key action
type (
	EncryptionKeyAction int64
	WindowAspect        int64
	WindowMode          int64
	WindowAction        int64
)

const (
	// WIDGET_NOTES_REFRESH_INTERVAL_MILLIS is the default refresh interval for the notes widget
	WIDGET_NOTES_REFRESH_INTERVAL_MILLIS = 2000 * time.Microsecond
	// Length of the encryption key generated by the application at first run
	ENCRYPTION_KEY_LENGTH = 256
	// STEF delete this
	// DEFAULT_ENCRYPTION_ALGORITHM = "aes-256-cbc"
	// CONFIG_ENCRYPTION_KEY               = "encryption_key"
	// CONFIG_ENCRYPTION_ALGORITHM         = "encryption_algorithm"
	CONFIG_CUR_ENCRYPTION_KEY_NAME      = "cur_encryption_key_name"
	CONFIG_ENCRYPTION_KEYS_PWD          = "encryption_keys_pwd"
	CONFIG_KVDB_PATH                    = "kvdb_path"
	CONFIG_GOOGLE_PROVIDER_PATH         = "google_provider_path"
	CONFIG_GOOGLE_CREDENTIALS_FILE_PATH = "google_credentials_file"
	CONFIG_GOOGLE_SHEET_ID              = "google_sheet_id"
	CONFIG_LOG_LEVEL                    = "log_level"
	CONFIG_LOG_FILE_PATH                = "log_file_path"
	CONFIG_KEY_FILE_PATH                = "key_file_path"

	EncryptionKeyAction_Generate EncryptionKeyAction = iota
	EncryptionKeyAction_Decrypt
	EncryptionKeyAction_Encrypt
	EncryptionKeyAction_Verify

	WindowAspect_Normal WindowAspect = iota
	WindowAspect_FullScreen

	WindowMode_Edit WindowMode = iota // edit/update mode
	WindowMode_View                   // read only

	WindowAction_New    WindowAction = iota // window with new data (create mode)
	WindowAction_Update                     // update data in window (update mode)
	WindowAction_Delete                     // prepare to delete data (delete mode)

	OPT_WINDOW_ASPECT = "window_aspect"
	OPT_WINDOW_MODE   = "window_mode"
	OPT_WINDOW_ACTION = "window_action"

	WIN_MAIN         = "main"
	WIN_NOTE_DETAILS = "note_details"

	BTN_CANCEL          = "btn_cancel"
	BTN_SAVE_NEW        = "btn_save_new"
	BTN_SAVE_UPDATED    = "btn_save_updated"
	BTN_DELETE          = "btn_delete"
	BTN_OK              = "btn_ok"
	BTN_TOGGLE_CONTENT  = "btn_toggle_content"
	BTN_COPY_ENCRYPTED  = "btn_copy_encrypted"
	BTN_PASTE_ENCRYPTED = "btn_paste_encrypted"
	BTN_PASSWORD_MODAL  = "btn_password_modal"

	WDG_NOTE_DETAILS_TITLE             = "note_details_title"
	WDG_NOTE_DETAILS_CONTENT           = "note_details_content"
	WDG_NOTE_DETAILS_CONTENT_RICH_TEXT = "note_details_content_rich_text"
	WDG_NOTE_DETAILS_HIDDEN            = "note_details_hidden"
	WDG_NOTE_DETAILS_ENCRYPTED         = "note_details_encrypted"
	WDG_NOTE_DETAILS_CREATED_AT        = "note_details_created_at"
	WDG_NOTE_DETAILS_UPDATED_AT        = "note_details_updated_at"
	WDG_NOTE_LIST                      = "note_list"
	WDG_PASSWORD_MODAL                 = "password_modal"
	WDG_SEARCH_BOX                     = "search_box"

	// log levels
	LOG_LEVEL_TRACE = "trace"
	LOG_LEVEL_DEBUG = "debug"
	LOG_LEVEL_INFO  = "info"
	LOG_LEVEL_WARN  = "warn"
	LOG_LEVEL_ERROR = "error"
	LOG_LEVEL_FATAL = "fatal"
	LOG_LEVEL_PANIC = "panic"

	ENCRYPTION_ALGORITHM_AES_256_CBC = "aes-256-cbc"
	ENCRYPTION_ALGORITHM_RSA_OAEP    = "rsa-oaep"
)

var (
	DEFAULT_RESOURCE_PATH           = "resources"
	DEFAULT_DB_PATH                 = "db/kv_store"
	DEFAULT_LOG_LEVEL               = LOG_LEVEL_ERROR
	DEFAULT_LOG_FILE_PATH           = filepath.Join("logs", "ecnotes.log")
	DEFAULT_KEY_FILE_PATH           = "key_store.json"
	SUPPORTED_ENCRYPTION_ALGORITHMS = []string{
		ENCRYPTION_ALGORITHM_AES_256_CBC,
		ENCRYPTION_ALGORITHM_RSA_OAEP,
	}
)
