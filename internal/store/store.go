package store

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"
)

type Store struct {
	mu       sync.RWMutex
	Data     map[string][]map[string]interface{}
	filePath string
	entities []string
}

func New(path string) (*Store, error) {
	s := &Store{
		Data:     make(map[string][]map[string]interface{}),
		filePath: path,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	for key, val := range raw {
		var items []map[string]interface{}
		if err := json.Unmarshal(val, &items); err == nil {
			s.Data[key] = items
			s.entities = append(s.entities, key)
		}
	}
	return s, nil
}

func NewEmpty(entities []string) *Store {
	s := &Store{
		Data: make(map[string][]map[string]interface{}),
	}
	for _, e := range entities {
		s.Data[e] = []map[string]interface{}{}
		s.entities = append(s.entities, e)
	}
	return s
}

func (s *Store) SetFilePath(path string) {
	s.filePath = path
}

func (s *Store) Entities() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.entities))
	copy(out, s.entities)
	return out
}

func (s *Store) List(entity string) []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items, ok := s.Data[entity]
	if !ok {
		return nil
	}
	out := make([]map[string]interface{}, len(items))
	for i, item := range items {
		out[i] = clone(item)
	}
	return out
}

func (s *Store) Get(entity, id string) (map[string]interface{}, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items, ok := s.Data[entity]
	if !ok {
		return nil, -1
	}
	for i, item := range items {
		if item["id"] == id {
			return clone(item), i
		}
	}
	return nil, -1
}

func (s *Store) Create(entity string, item map[string]interface{}) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := item["id"]; !ok || item["id"] == nil || item["id"] == "" {
		item["id"] = uuid.New().String()
	}
	item = clone(item)
	s.Data[entity] = append(s.Data[entity], item)
	s.persist()
	return item
}

func (s *Store) Update(entity string, idx int, item map[string]interface{}) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	item["id"] = s.Data[entity][idx]["id"]
	item = clone(item)
	s.Data[entity][idx] = item
	s.persist()
	return item
}

func (s *Store) Patch(entity string, idx int, patch map[string]interface{}) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range patch {
		if k != "id" {
			s.Data[entity][idx][k] = v
		}
	}
	s.persist()
	return clone(s.Data[entity][idx])
}

func (s *Store) Delete(entity string, idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data[entity] = append(s.Data[entity][:idx], s.Data[entity][idx+1:]...)
	s.persist()
}

func (s *Store) persist() {
	if s.filePath == "" {
		return
	}
	raw := make(map[string]interface{})
	for _, entity := range s.entities {
		raw[entity] = s.Data[entity]
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.filePath, data, 0644)
}

func clone(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
