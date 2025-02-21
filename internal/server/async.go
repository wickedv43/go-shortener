package server

import (
	"fmt"
	"sync"
)

func (s *Server) gen(shorts ...string) chan string {
	outCh := make(chan string)
	go func() {
		defer close(outCh)
		for _, short := range shorts {
			outCh <- short
		}
	}()

	return outCh
}

func (s *Server) delete(inCh chan string) chan string {
	outCh := make(chan string)
	go func() {
		defer close(outCh)
		for short := range inCh {
			err := s.batchDelete(short)
			if err != nil {
				s.logger.Error(err)
			}
			outCh <- fmt.Sprintf("deleted: %s", short)
		}
	}()

	return outCh
}

func (s *Server) fanIn(chs ...chan string) chan string {
	var wg sync.WaitGroup
	outCh := make(chan string)

	output := func(c chan string) {
		for n := range c {
			outCh <- n
		}
		wg.Done()
	}

	wg.Add(len(chs))

	for _, c := range chs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
