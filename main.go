package main

import (
	"fmt"
	"math"
	"sort"
)

const numCards = 4
const maxCard = 10.0
const targetNum = 24.0

var operators = []operator{AddOp, SubOp, MulOp, DivOp}
var treeCache = make(map[string][]node)
var neededTrees = make(map[string]node)

func main() {
	possible := 0

	all := allDeals([]float64{})

	for i, deal := range all {
		tree := getNeededTree(deal, targetNum)
		if tree != nil {
			possible++
			fmt.Printf("%d. %v: %s\n", i, deal, tree.getExp())
		} else {
			fmt.Printf("%d. %v: not possible\n", i, deal)
		}
	}

	fmt.Printf("%d / %d possible\n", possible, len(all))
}

func allDeals(partialDeal []float64) [][]float64 {
	if len(partialDeal) == numCards {
		return [][]float64{partialDeal}
	}
	lastCard := 1.0
	if len(partialDeal) > 0 {
		lastCard = partialDeal[len(partialDeal)-1]
	}
	var all [][]float64
	for i := lastCard; i <= maxCard; i++ {
		newPartial := append([]float64{}, partialDeal...)
		newPartial = append(newPartial, i)
		all = append(all, allDeals(newPartial)...)
	}
	return all
}

func getPossibleTrees(deal []float64) []node {
	ds := dealString(deal)
	if nodes, ok := treeCache[ds]; ok {
		return nodes
	}
	if len(deal) == 1 {
		return []node{&valNode{deal[0]}}
	}
	var trees []node
	for i := 1; i < 1<<uint64(len(deal))-1; i++ {
		// the bits in i determine which elements go right vs. left.
		// 0 means left, 1 means left.
		// we start from 1 and go to 2^N - 2 to skip the 0...0 and 1...1 cases.
		leftCards, rightCards := splitDeal(deal, i)
		leftTrees := getPossibleTrees(leftCards)
		rightTrees := getPossibleTrees(rightCards)
		for _, lt := range leftTrees {
			for _, rt := range rightTrees {
				for _, op := range operators {
					trees = append(trees, &opNode{op, lt, rt})
				}
			}
		}
	}
	for _, tree := range trees {
		cacheString := fmt.Sprintf("%s:%v", dealString(deal), tree.getVal())
		neededTrees[cacheString] = tree
	}
	treeCache[ds] = trees
	return trees
}

func getNeededTree(deal []float64, neededVal float64) node {
	cacheString := fmt.Sprintf("%s:%v", dealString(deal), neededVal)
	if tree, ok := neededTrees[cacheString]; ok {
		return tree
	}
	if len(deal) == 1 {
		if closeEnough(deal[0], neededVal) {
			return &valNode{deal[0]}
		}
		return nil
	}
	for i := 1; i < 1<<uint64(len(deal))-1; i++ {
		leftCards, rightCards := splitDeal(deal, i)
		for _, lt := range getPossibleTrees(leftCards) {
			leftVal := lt.getVal()
			for _, op := range operators {
				wantedVal := op.computeNecessary(leftVal, neededVal)
				if op == DivOp && wantedVal == 0 {
					continue
				}
				if rt := getNeededTree(rightCards, wantedVal); rt != nil {
					t := &opNode{op, lt, rt}
					neededTrees[cacheString] = t
					return t
				}
			}
		}
	}
	neededTrees[cacheString] = nil
	return nil
}

func closeEnough(a, b float64) bool {
	return math.Abs(a-b) < 0.00001
}

func dealString(deal []float64) string {
	sortedDeal := append([]float64{}, deal...)
	sort.Sort(sort.Float64Slice(sortedDeal))
	return fmt.Sprintf("%v", sortedDeal)
}

// splitDeal splits the provided deal based on the bits in deal.
func splitDeal(deal []float64, split int) ([]float64, []float64) {
	var leftCards []float64
	var rightCards []float64
	for bit := 0; bit < len(deal); bit++ {
		if (split>>uint64(bit))%2 == 0 {
			leftCards = append(leftCards, deal[bit])
		} else {
			rightCards = append(rightCards, deal[bit])
		}
	}
	return leftCards, rightCards
}

type node interface {
	getVal() float64
	getExp() string
}

// valNode is a value node in the expression tree.
type valNode struct {
	val float64
}

func (v *valNode) getVal() float64 { return v.val }

func (v *valNode) getExp() string { return fmt.Sprintf("%v", v.val) }

// opNode is an binary operator node in the expression tree.
type opNode struct {
	op    operator
	left  node
	right node
}

func (o *opNode) getVal() float64 { return o.op.compute(o.left.getVal(), o.right.getVal()) }

func (o *opNode) getExp() string {
	return fmt.Sprintf("(%s %s %s)", o.left.getExp(), o.op.opString(), o.right.getExp())
}

type operator interface {
	compute(left, right float64) float64
	computeNecessary(left, right float64) float64
	opString() string
}

type addOp struct{}

func (*addOp) compute(left, right float64) float64 { return left + right }

func (*addOp) computeNecessary(left, want float64) float64 { return want - left }

func (*addOp) opString() string { return "+" }

var AddOp = &addOp{}

type subOp struct{}

func (*subOp) compute(left, right float64) float64 { return left - right }

func (*subOp) computeNecessary(left, want float64) float64 { return left - want }

func (*subOp) opString() string { return "-" }

var SubOp = &subOp{}

type mulOp struct{}

func (*mulOp) compute(left, right float64) float64 { return left * right }

func (*mulOp) computeNecessary(left, want float64) float64 { return want / left }

func (*mulOp) opString() string { return "*" }

var MulOp = &mulOp{}

type divOp struct{}

func (*divOp) compute(left, right float64) float64 { return left / right }

func (*divOp) computeNecessary(left, want float64) float64 { return left / want }

func (*divOp) opString() string { return "/" }

var DivOp = &divOp{}
