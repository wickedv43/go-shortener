package main

type Storage struct {
	Data map[string]string `json:"data"`
}

func (s *Storage) Init() {
	s.Data = make(map[string]string)
}

func (s *Storage) Save(url string, short string) {
	s.Data[short] = url
}

func (s *Storage) Get(short string) (string, bool) {
	url, ok := s.Data[short]
	return url, ok
}

func (s *Storage) InStorage(url string) (bool, string) {
	var res bool
	for short, v := range s.Data {
		if v == url {
			return true, short
		}
	}
	return res, ""
}

var S Storage
