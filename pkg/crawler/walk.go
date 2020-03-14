package crawler

import (
	"golang.org/x/net/html"
)

type WalkFunc func(node *html.Node) error

type Walker struct {
	Func WalkFunc
}

func (w Walker) Walk(root *html.Node) error {
	var walk func(node *html.Node) error
	walk = func(node *html.Node) error {
		for n := node.FirstChild; n != nil; n = n.NextSibling {
			if w.Func != nil {
				err := w.Func(n)
				if err != nil {
					return err
				}
			}
			err := walk(n)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return walk(root)
}

type TransformFunc func(node *html.Node) (*html.Node, error)

type Transformer struct {
	Func TransformFunc
}

func (w Transformer) WalkTransform(root *html.Node) error {
	var walk func(node *html.Node) error
	walk = func(node *html.Node) error {
		for n := node.FirstChild; n != nil; {
			// copy next node before removing the node n
			next := n.NextSibling

			if w.Func != nil {
				newChild, err := w.Func(n)
				if err != nil {
					return err
				}
				if newChild == nil {
					node.RemoveChild(n)
				} else if newChild != n {
					node.InsertBefore(newChild, n)
					node.RemoveChild(n)
				}
			}
			err := walk(n)
			if err != nil {
				return err
			}
			n = next
		}
		return nil
	}
	return walk(root)
}
