/**
 * This file contains an implementation of a priority queue for the supervisor.
 *
 * @author Jacob Bringham, Raleigh Martin, Parth Parikh
 * @version 4/17/2022
 **/

package main

import (
	"container/heap"

	"github.com/showalter/bdws/internal/data"
)

// -- Internal Structs --------------------------------------------------------

type Item struct {
	index    int
	priority int
	job      data.Job
}

type PriorityQueue []*Item

// -- Internal Routines  ------------------------------------------------------

/** -- Len() ------------------------------------------------------------------
 *  Returns the size of the priority queue.
 ** ------------------------------------------------------------------------ */
func (pq PriorityQueue) Len() int { return len(pq) }

/** -- compare() --------------------------------------------------------------
 *  Compares two Items within the priority queue and returns true if item i is
 *  greater than item j.
 *
 *  @param i Index of an item in the queue
 *  @param j Index of an item in the queue
 *  @return TRUE if i > j
 ** ------------------------------------------------------------------------ */
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

/** -- Swap() -----------------------------------------------------------------
* Swaps two Items in the queue.
*
* @param i Index of first item
* @param j Index of second item
** ------------------------------------------------------------------------ */
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

/** -- pushItem() -------------------------------------------------------------
 *  Puts an item into the priority queue.
 *
 *  @param x item to place into priority queue
 ** ------------------------------------------------------------------------ */
func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

/** -- popItem() --------------------------------------------------------------
 *  Removes the highest priority item from the queue.
 *
 *  @return Highest priority item in the queue
 ** ------------------------------------------------------------------------ */
func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

/** -- update() ---------------------------------------------------------------
 * Modifies the priority and value of an Item in the queue.
 *
 * @param item     Item to update
 * @param priority New priority of item
 ** ------------------------------------------------------------------------ */
func (pq *PriorityQueue) update(item *Item, priority int) {
	item.priority = priority
	heap.Fix(pq, item.index)
}
