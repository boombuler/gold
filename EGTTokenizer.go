package gold

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

func (g *egtGrammar) readTokens(rd io.Reader) <-chan *parserToken {
	c := make(chan *parserToken)
	r := newSourceReader(rd)
	go func() {
		for {
			t := g.readToken(r, true)
			c <- t
			if t.Symbol.Kind == stEnd {
				close(c)
			}
		}
	}()
	return g.processGroups(c)
}

func (g *egtGrammar) processGroups(input <-chan *parserToken) <-chan *parserToken {
	result := make(chan *parserToken)

	groupStack := new(stack)

	go func() {

		nestGroup := false

		for read := range input {

			fmt.Println("Processing ", read)

			// Groups (comments, etc.)
			// The logic - to determine if a group should be nested - requires that the top
			// of the stack and the symbol's linked group need to be looked at. Both of these
			// can be unset. So, this section sets a boolean and avoids errors. We will use
			// this boolean in the logic chain below.
			if read.Symbol.Kind == stGroupStart {
				if groupStack.Len() == 0 {
					nestGroup = true
				} else {
					nestGroup = groupStack.Peek().(*parserToken).Symbol.Group.Nested.contains(read.Symbol.Group)
				}
			} else {
				nestGroup = false
			}

			// Logic chain
			if nestGroup {
				groupStack.Push(read)
			} else if groupStack.Len() == 0 {
				// The token is ready to be analyzed
				result <- read
			} else if groupStack.Peek().(*parserToken).Symbol.Group.End == read.Symbol {
				// End the current group
				pop := groupStack.Pop().(*parserToken)

				// Ending logic
				if pop.Symbol.Group.EndingMode == emClosed {
					pop.Text = pop.Text + read.Text
				}

				if groupStack.Len() == 0 {
					// We are out of the group. Return pop'd token which contains all the group text
					pop.Symbol = pop.Symbol.Group.Container
					result <- read
				} else {
					// Append group text to parent
					groupStack.Peek().(*parserToken).Text += pop.Text
				}
			} else if read.Symbol.Kind == stEnd {
				// EOF always stops the loop. The caller method (parse) can flag a runaway group error.
				result <- read
				break
			} else {
				// We are in a group, Append to the Token on the top of the stack.
				// Take into account the Token group mode
				top := groupStack.Peek().(*parserToken)
				if top.Symbol.Group.AdvanceMode == amToken {
					// Append all text
					top.Text += read.Text
				} else {
					// Append one character
					top.Text += string([]rune(read.Text)[0])
					//! consumeBuffer(1);
				}
			}
		}
		close(result)
	}()

	return result
}

func (g *egtGrammar) readToken(r *sourceReader, readComments bool) *parserToken {

	dfa := g.getInitialDfaState()

	tText := new(bytes.Buffer)
	tWriter := bufio.NewWriter(tText)

	result := new(parserToken)
	result.Text = ""
	result.Symbol = g.errorSymbol
	result.Position = r.Position

	for {
		if !r.Next() {
			tWriter.Flush()
			if r.Rune == 0 && tText.Len() == 0 {
				result.Text = ""
				result.Symbol = g.endSymbol
			}
			return result
		}

		nextState, ok := dfa.TransitionVector(r.Rune)
		if ok {
			tWriter.WriteRune(r.Rune)

			dfa = nextState
			if dfa.AcceptSymbol != nil {
				tWriter.Flush()
				result.Text = string(tText.Bytes())
				result.Symbol = dfa.AcceptSymbol
			}
		} else {
			if result.Symbol == g.errorSymbol {
				tWriter.WriteRune(r.Rune)
			}

			r.SkipNextRead = true
			break
		}
	}

	tWriter.Flush()
	result.Text = string(tText.Bytes())

	return result
}
