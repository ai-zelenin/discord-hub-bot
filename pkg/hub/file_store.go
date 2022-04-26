package hub

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type FileStore struct {
	data   map[string]map[string]*Subscription
	mx     sync.RWMutex
	fn     string
	logger *log.Logger
}

func NewFileStore(logger *log.Logger, fn string) (*FileStore, error) {
	_, err := os.Stat(fn)
	m := make(map[string]map[string]*Subscription)
	if err == nil {
		data, err := ioutil.ReadFile(fn)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &m)
		if err != nil {
			return nil, err
		}
	}

	s := &FileStore{
		data:   m,
		fn:     fn,
		logger: logger,
	}
	go s.SaveLoop()
	return s, nil
}

func (s *FileStore) SaveLoop() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for range ticker.C {
		s.Save()
	}
}

func (s *FileStore) Save() {
	s.mx.RLock()
	defer s.mx.RUnlock()
	data, err := json.MarshalIndent(s.data, "", "\t")
	if err != nil {
		s.logger.Println(err)
	}
	err = ioutil.WriteFile(s.fn, data, os.ModePerm)
	if err != nil {
		s.logger.Println(err)
	}
}

func (s *FileStore) FindSubscriptionsBySource(source string) ([]*Subscription, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	subMap := s.data[source]
	subs := make([]*Subscription, 0)
	for _, sub := range subMap {
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *FileStore) AddSubscription(sub *Subscription) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	subs, ok := s.data[sub.Source]
	if !ok {
		subs = make(map[string]*Subscription, 0)
	}
	subs[sub.UserID] = sub
	s.data[sub.Source] = subs
	return nil
}

func (s *FileStore) RemoveSubscription(source string, userID string) (*Subscription, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	subMap, ok := s.data[source]
	if !ok {
		return nil, nil
	}
	sub, ok := subMap[userID]
	if !ok {
		return nil, nil
	}
	delete(subMap, userID)
	return sub, nil
}
