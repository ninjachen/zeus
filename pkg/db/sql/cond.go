package sql

import "xorm.io/xorm"

type Cond interface {
	Apply(s *xorm.Session) error
}

type Pagination struct {
	size int32
	page int32
}

type Limit struct {
	size   int32
	offset int32
}

type OrderBy struct {
	order string
}

type GroupBy struct {
	group string
}

type CondEx struct {
	f func(s *xorm.Session) error
}

func (self *CondEx) Apply(s *xorm.Session) error {
	if self.f != nil {
		return self.f(s)
	}
	return nil
}

func (self *OrderBy) Apply(s *xorm.Session) error {
	s.OrderBy(self.order)
	return nil
}

func (self *Pagination) Apply(s *xorm.Session) error {
	s.Limit(int(self.size), int(self.size*(self.page-1)))
	return nil
}

func (self *Limit) Apply(s *xorm.Session) error {
	s.Limit(int(self.size), int(self.offset))
	return nil
}

func (self *GroupBy) Apply(s *xorm.Session) error {
	s.GroupBy(self.group)
	return nil
}

func NewPagination(size, page int32) *Pagination {
	return &Pagination{size: size, page: page}
}

func NewOrderBy(order string) *OrderBy {
	return &OrderBy{order: order}
}

func NewLimit(size, offset int32) *Limit {
	return &Limit{
		size:   size,
		offset: offset,
	}
}

func NewGroupBy(group string) *GroupBy {
	return &GroupBy{group: group}
}

func NewCondEx(f func(s *xorm.Session) error) *CondEx {
	return &CondEx{f: f}
}
