// Copyright 2024 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memo

import (
	base2 "github.com/pingcap/tidb/pkg/planner/cascades/base"
	"github.com/pingcap/tidb/pkg/planner/cascades/pattern"
	"github.com/pingcap/tidb/pkg/planner/cascades/util"
	"github.com/pingcap/tidb/pkg/planner/core/base"
	"github.com/pingcap/tidb/pkg/util/intest"
)

// GroupExpression is a single expression from the equivalent list classes inside a group.
// it is a node in the expression tree, while it takes groups as inputs. This kind of loose
// coupling between Group and GroupExpression is the key to the success of the memory compact
// of representing a forest.
type GroupExpression struct {
	// LogicalPlan is internal logical expression stands for this groupExpr.
	// Define it in the header element can make GE as Logical Plan implementor.
	base.LogicalPlan

	// group is the Group that this GroupExpression belongs to.
	group *Group

	// inputs stores the Groups that this GroupExpression based on.
	Inputs []*Group

	// hash64 is the unique fingerprint of the GroupExpression.
	hash64 uint64
}

// GetGroup returns the Group that this GroupExpression belongs to.
func (e *GroupExpression) GetGroup() *Group {
	return e.group
}

// String implements the fmt.Stringer interface.
func (e *GroupExpression) String(w util.StrBufferWriter) {
	e.LogicalPlan.ExplainID()
	w.WriteString("GE:" + e.LogicalPlan.ExplainID().String() + "{")
	for i, input := range e.Inputs {
		if i != 0 {
			w.WriteString(", ")
		}
		input.String(w)
	}
	w.WriteString("}")
}

// GetHash64 returns the cached hash64 of the GroupExpression.
func (e *GroupExpression) GetHash64() uint64 {
	intest.Assert(e.hash64 != 0, "hash64 should not be 0")
	return e.hash64
}

// Hash64 implements the Hash64 interface.
func (e *GroupExpression) Hash64(h base2.Hasher) {
	// logical plan hash.
	e.LogicalPlan.Hash64(h)
	// children group hash.
	for _, child := range e.Inputs {
		child.Hash64(h)
	}
}

// Equals implements the Equals interface.
func (e *GroupExpression) Equals(other any) bool {
	e2, ok := other.(*GroupExpression)
	if !ok {
		return false
	}
	if e == nil {
		return e2 == nil
	}
	if e2 == nil {
		return false
	}
	if len(e.Inputs) != len(e2.Inputs) {
		return false
	}
	if pattern.GetOperand(e.LogicalPlan) != pattern.GetOperand(e2.LogicalPlan) {
		return false
	}
	// current logical operator meta cmp, logical plan don't care logicalPlan's children.
	// when we convert logicalPlan to GroupExpression, we will set children to nil.
	if !e.LogicalPlan.Equals(e2.LogicalPlan) {
		return false
	}
	// if one of the children is different, then the two GroupExpressions are different.
	for i, one := range e.Inputs {
		if !one.Equals(e2.Inputs[i]) {
			return false
		}
	}
	return true
}

// NewGroupExpression creates a new GroupExpression with the given logical plan and children.
func NewGroupExpression(lp base.LogicalPlan, inputs []*Group) *GroupExpression {
	return &GroupExpression{
		group:       nil,
		Inputs:      inputs,
		LogicalPlan: lp,
		hash64:      0,
	}
}

// Init initializes the GroupExpression with the given group and hasher.
func (e *GroupExpression) Init(h base2.Hasher) {
	e.Hash64(h)
	e.hash64 = h.Sum64()
}
