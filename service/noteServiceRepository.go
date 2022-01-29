package service

// NoteServiceRepository ....
type NoteServiceRepository interface {
	GetAllNotes() ([]Note, error)
	GetNote(id int) (Note, error)
	CreateNote(note Note) error
	UpdateNote(note Note) error
	DeleteNote(id int) error
}

// NoteServiceRepositoryImpl ....
type NoteServiceRepositoryImpl struct{}

// NewNoteServiceRepository ....
func NewNoteServiceRepository() NoteServiceRepository {
	return &NoteServiceRepositoryImpl{}
}

// GetAllNotes ....
func (nsr *NoteServiceRepositoryImpl) GetAllNotes() ([]Note, error) {
	panic("not implemented") // TODO: Implement
}

// GetNote ....
func (nsr *NoteServiceRepositoryImpl) GetNote(id int) (Note, error) {
	panic("not implemented") // TODO: Implement
}

// CreateNote ....
func (nsr *NoteServiceRepositoryImpl) CreateNote(note Note) error {
	panic("not implemented") // TODO: Implement
}

// UpdateNote ....
func (nsr *NoteServiceRepositoryImpl) UpdateNote(note Note) error {
	panic("not implemented") // TODO: Implement
}

// DeleteNote ....
func (nsr *NoteServiceRepositoryImpl) DeleteNote(id int) error {
	panic("not implemented") // TODO: Implement
}
