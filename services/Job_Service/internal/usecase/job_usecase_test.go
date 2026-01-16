package usecase

import (
	"job_service/internal/domain"
	"testing"
)

type mockKafka struct {
	called bool 
}

func (m *mockKafka) Publish(topic, key string, value []byte) error {
	m.called = true
	return nil
}

func TestCreateJob(t *testing.T) {
	k := &mockKafka{}
	uc := NewJobUsecase(k, nil)

	err := uc.CreateJob(nil, domain.Job{JobID: "1"}, "idem1")
	if err != nil {
		t.Fatal(err)
	}

	if !k.called {
		t.Fatal("kafka not called")
	}
}