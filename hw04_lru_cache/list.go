package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

func (list *list) MoveToFront(i *ListItem) {
	if i != list.front {
		// front = i
		// у бывшего первым предыдущий = новый первый
		// у следующего предыдущий = i.Prev (если есть следующий)
		// у предыдущего следующий = i.Next (если есть предыдущий)
		// у нового первого предыдущий = nil, следующий = бывшему первому
		// если двигаем последний, то back = предпоследний

		if list.front != nil {
			list.front.Prev = i
		}
		if i.Next != nil {
			i.Next.Prev = i.Prev
		}
		if i.Prev != nil {
			i.Prev.Next = i.Next
		}
		if i == list.back {
			list.back = i.Prev
		}
		i.Prev = nil
		i.Next = list.front
		list.front = i
	}
}

func (list *list) Remove(i *ListItem) {
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	}
	if list.front == i {
		list.front = i.Next
	}
	if list.back == i {
		list.back = i.Prev
	}
	list.len--
}

func (list list) Front() *ListItem {
	return list.front
}

func (list list) Back() *ListItem {
	return list.back
}

func (list *list) PushFront(v interface{}) *ListItem {
	front := new(ListItem)
	front.Value = v
	front.Next = list.front
	list.front = front
	if list.len == 0 {
		list.back = front
	}
	if front.Next != nil {
		front.Next.Prev = front
	}
	list.len++
	return front
}

func (list *list) PushBack(v interface{}) *ListItem {
	back := new(ListItem)
	back.Value = v
	back.Prev = list.back
	list.back = back
	if list.len == 0 {
		list.front = back
	}
	if back.Prev != nil {
		back.Prev.Next = back
	}
	list.len++
	return back
}

func (list list) Len() int {
	return list.len
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
}

func NewList() List {
	return new(list)
}
