package storage

func (s *Storage) Put(url string, short string) {
	s.db[short] = url
}

func (s *Storage) Get(short string) (string, bool) {
	url, ok := s.db[short]
	return url, ok
}

func (s *Storage) InStorage(url string) (bool, string) {
	var res bool
	for short, v := range s.db {
		if v == url {
			return true, short
		}
	}
	return res, ""
}
