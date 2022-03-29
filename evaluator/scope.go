package evaluator

type Scope struct {
	store       map[string]Object
	parentScope *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{store: map[string]Object{}, parentScope: parent}
}

func (s *Scope) Get(key string) (Object, bool) {
	scope := s
	for scope.store[key] == nil && scope.parentScope != nil {
		scope = scope.parentScope
	}
	v, ok := scope.store[key]
	return v, ok
}
func (s *Scope) Delete(key string) (Object, bool) {
	scope := s
	for scope.store[key] == nil && scope.parentScope != nil {
		scope = scope.parentScope
	}
	if v, ok := scope.store[key]; ok {
		delete(scope.store, key)
		return v, ok
	}
	return nil, false
}
func (s *Scope) GetCurrent(key string) (Object, bool) {
	v, ok := s.store[key]
	return v, ok
}
func (s *Scope) Set(key string, value Object) Object {
	s.store[key] = value
	return value
}
func (s *Scope) Reset(key string, value Object) (Object, bool) {
	scope := s
	for scope.store[key] == nil && scope.parentScope != nil {
		scope = scope.parentScope
	}
	if _, ok := scope.store[key]; ok {
		scope.store[key] = value
		return value, true
	}
	return value, false
}
