package server

func (s *Server) deleteWorker(inCh chan string) chan string {
	outCh := make(chan string)
	go func() {
		defer close(outCh)
		for short := range inCh {
			err := s.batchDelete(short)
			if err != nil {
				s.logger.Errorf("delete error: %v", err)
			}
		}
	}()
	return outCh
}
