package errors

import (
	"errors"
	"fmt"
)

// MessageCountMap 包含每个错误消息的出现次数.
type MessageCountMap map[string]int

// Aggregate 聚合表示一个包含多个错误的对象，但不一定具有单一的语义.
// 聚合可以与“errors.Is（）”一起使用，以检查是否出现特定的错误类型.
// 不支持Errors.As（），因为调用者可能关心与给定类型匹配的潜在多个特定错误
type Aggregate interface {
	error
	Errors() []error
	Is(error) bool
}

// NewAggregate 将一段错误转换为Aggregate接口，该接口本身就是错误接口的实现.如果切片为空，则返回nil.
// 它将检查输入错误列表中的任何元素是否为nil，以避免调用error（）时nil指针死机.
func NewAggregate(errlist []error) Aggregate {
	if len(errlist) == 0 {
		return nil
	}
	// 如果输入错误列表包含nil
	var errs []error
	for _, e := range errlist {
		if e != nil {
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return aggregate(errs)
}

// 这个助手实现了error和Errors接口.将其保持为私有可以防止人们生成0个错误的聚合，这不是错误，但确实满足错误接口.
type aggregate []error

// Error 是 error interface 的一部分.
func (agg aggregate) Error() string {
	if len(agg) == 0 {
		// This should never happen, really.
		return ""
	}
	if len(agg) == 1 {
		return agg[0].Error()
	}
	seenerrs := NewString()
	result := ""
	agg.visit(func(err error) bool {
		msg := err.Error()
		if seenerrs.Has(msg) {
			return false
		}
		seenerrs.Insert(msg)
		if len(seenerrs) > 1 {
			result += ", "
		}
		result += msg
		return false
	})
	if len(seenerrs) == 1 {
		return result
	}
	return "[" + result + "]"
}

func (agg aggregate) Is(target error) bool {
	return agg.visit(func(err error) bool {
		return errors.Is(err, target)
	})
}

func (agg aggregate) visit(f func(err error) bool) bool {
	for _, err := range agg {
		switch err := err.(type) {
		case aggregate:
			if match := err.visit(f); match {
				return match
			}
		case Aggregate:
			for _, nestedErr := range err.Errors() {
				if match := f(nestedErr); match {
					return match
				}
			}
		default:
			if match := f(err); match {
				return match
			}
		}
	}

	return false
}

// Errors 是 Aggregate interface 的一部分.
func (agg aggregate) Errors() []error {
	return []error(agg)
}

// Matcher 用于匹配错误.如果错误匹配，则返回true.
type Matcher func(error) bool

// FilterOut 从输入错误中删除与任何匹配器匹配的所有错误.
// 如果输入是一个奇异误差，则只测试该误差.
// 如果输入实现了Aggregate接口，那么错误列表将被递归处理.
// 例如，这可以用于从错误列表中删除已知的OK错误（如io.EOF或os.PathNotFound）.
func FilterOut(err error, fns ...Matcher) error {
	if err == nil {
		return nil
	}
	if agg, ok := err.(Aggregate); ok {
		return NewAggregate(filterErrors(agg.Errors(), fns...))
	}
	if !matchesError(err, fns...) {
		return err
	}
	return nil
}

// matchesError 返回true 如果任何 Matcher 返回true
func matchesError(err error, fns ...Matcher) bool {
	for _, fn := range fns {
		if fn(err) {
			return true
		}
	}
	return false
}

// filterErrors 返回所有fns都返回false的任何错误（或嵌套错误，如果列表中包含嵌套错误）.
// 如果没有错误，则返回一个nil列表.产生的silec将使所有嵌套切片变平，作为副作用.
func filterErrors(list []error, fns ...Matcher) []error {
	result := []error{}
	for _, err := range list {
		r := FilterOut(err, fns...)
		if r != nil {
			result = append(result, r)
		}
	}
	return result
}

// Flatten 获取一个聚合，该聚合可以将其他聚合保存在任意嵌套中，并递归地将它们全部展平为一个聚合
func Flatten(agg Aggregate) Aggregate {
	result := []error{}
	if agg == nil {
		return nil
	}
	for _, err := range agg.Errors() {
		if a, ok := err.(Aggregate); ok {
			r := Flatten(a)
			if r != nil {
				result = append(result, r.Errors()...)
			}
		} else {
			if err != nil {
				result = append(result, err)
			}
		}
	}
	return NewAggregate(result)
}

// CreateAggregateFromMessageCountMap 转换 MessageCountMap 聚合
func CreateAggregateFromMessageCountMap(m MessageCountMap) Aggregate {
	if m == nil {
		return nil
	}
	result := make([]error, 0, len(m))
	for errStr, count := range m {
		var countStr string
		if count > 1 {
			countStr = fmt.Sprintf(" (repeated %v times)", count)
		}
		result = append(result, fmt.Errorf("%v%v", errStr, countStr))
	}
	return NewAggregate(result)
}

// Reduce 将返回err，或者，如果err是一个聚合并且只有一个项，则返回聚合中的第一个项.
func Reduce(err error) error {
	if agg, ok := err.(Aggregate); ok && err != nil {
		switch len(agg.Errors()) {
		case 1:
			return agg.Errors()[0]
		case 0:
			return nil
		}
	}
	return err
}

// AggregateGoroutines 并行运行所提供的函数，将所有非零错误填充到返回的Aggregate中.
// 如果所有函数都成功完成，则返回nil.
func AggregateGoroutines(funcs ...func() error) Aggregate {
	errChan := make(chan error, len(funcs))
	for _, f := range funcs {
		go func(f func() error) { errChan <- f() }(f)
	}
	errs := make([]error, 0)
	for i := 0; i < cap(errChan); i++ {
		if err := <-errChan; err != nil {
			errs = append(errs, err)
		}
	}
	return NewAggregate(errs)
}

// ErrPreconditionViolated 在违反前提条件时返回
var ErrPreconditionViolated = errors.New("precondition is violated")
