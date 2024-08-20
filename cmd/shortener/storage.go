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

func (s *Storage) Get(short string) string {
	return s.Data[short]
}

func (s *Storage) InStorage(url string) bool {
	var res bool
	for _, v := range s.Data {
		if v == url {
			res = true
		}
	}
	return res
}

var S Storage
