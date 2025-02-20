package server

import (
	"time"
)

const batchSize = 10
const flushInterval = 2 * time.Second

func (s *Server) deleteWorker() {

	batch := make([]string, 0, batchSize)
	timer := time.NewTimer(flushInterval)

	for {
		select {
		case short := <-s.urlDeleteChan:

			batch = append(batch, short)

			// Если накопили batchSize, отправляем в БД
			if len(batch) >= batchSize {
				s.batchDelete(batch)
				batch = make([]string, 0, batchSize)
			}

		case <-timer.C:
			// Если время вышло, но есть данные, тоже отправляем
			if len(batch) > 0 {
				s.batchDelete(batch)
				batch = make([]string, 0, batchSize)
			}
			timer.Reset(flushInterval)
		}
	}
}
