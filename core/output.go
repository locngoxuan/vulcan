package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var tmpDir = filepath.Join("/tmp", "vulcan")
var variableDSN = filepath.Join(tmpDir, "database")

func CreateTmpDir() error {
	return os.MkdirAll(tmpDir, 0755)
}

func SetCurrentStep(stepId string) error {
	db, err := bolt.Open(variableDSN, 0755, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf(`failed to open db. %v`, err)
	}
	defer func() {
		_ = db.Close()
	}()

	stepId = strings.TrimSpace(stepId)
	if stepId == "" {
		return nil
	}

	return db.Update(func(t *bolt.Tx) (err error) {
		bck := t.Bucket([]byte("current_step"))
		if bck == nil {
			bck, err = t.CreateBucketIfNotExists([]byte("current_step"))
			if err != nil {
				return
			}
		}
		return bck.Put([]byte("value"), []byte(stepId))
	})
}

func SetOutput(key, value string) error {
	db, err := bolt.Open(variableDSN, 0755, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf(`failed to open db. %v`, err)
	}
	defer func() {
		_ = db.Close()
	}()

	var stepId string = ""
	err = db.View(func(t *bolt.Tx) error {
		bck := t.Bucket([]byte("current_step"))
		bytes := bck.Get([]byte("value"))
		if bytes == nil || len(bytes) == 0 {
			return nil
		}
		stepId = string(bytes)
		return nil
	})
	if err != nil {
		return err
	}

	stepId = strings.TrimSpace(stepId)
	if stepId == "" {
		return nil
	}

	return db.Update(func(t *bolt.Tx) (err error) {
		bck := t.Bucket([]byte("variables"))
		if bck == nil {
			bck, err = t.CreateBucketIfNotExists([]byte("variables"))
			if err != nil {
				return err
			}
		}
		key = fmt.Sprintf(`steps_%s_outputs_%s`, stepId, key)
		return bck.Put([]byte(key), []byte(value))
	})
}

type OutputVal struct {
	Key   string
	Value string
}

func GetAllOutputs() ([]OutputVal, error) {
	var outputs []OutputVal
	db, err := bolt.Open(variableDSN, 0755, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return outputs, fmt.Errorf(`failed to open db. %v`, err)
	}
	defer func() {
		_ = db.Close()
	}()

	err = db.View(func(t *bolt.Tx) error {
		bck := t.Bucket([]byte("variables"))
		if bck == nil {
			return nil
		}
		bck.ForEach(func(k, v []byte) error {
			outputs = append(outputs, OutputVal{
				Key:   string(k),
				Value: string(v),
			})
			return nil
		})
		return nil
	})

	return outputs, err
}
