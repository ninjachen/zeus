package sql

type ExecResult struct {
	err error
}

type FindResult struct {
	ExecResult
	count int32
}

type AffectedResult struct {
	ExecResult
	affected int32
}

type InsertResult struct {
	ExecResult
	affected int32
}

func NewFindResult(count int64, err error) *FindResult {
	return &FindResult{ExecResult{err}, int32(count)}
}

func NewAffectedResult(count int64, err error) *AffectedResult {
	return &AffectedResult{ExecResult{err}, int32(count)}
}

func NewInsertResult(count int64, err error) *InsertResult {
	return &InsertResult{ExecResult{err}, int32(count)}
}

func (self *ExecResult) HasError() bool {
	return self.err != nil
}

func (self *ExecResult) Error() error {
	return self.err
}

func (self *AffectedResult) Affected() int32 {
	return self.affected
}

func (self *InsertResult) Success() bool {
	return self.affected > 0 && self.err == nil
}

func (self *AffectedResult) Success() bool {
	return self.err == nil
}

func (self *FindResult) NotFound() bool {
	return self.count == 0
}

func (self *FindResult) HasFound() bool {
	return self.count != 0
}

func (self *FindResult) Count() int32 {
	return int32(self.count)
}
