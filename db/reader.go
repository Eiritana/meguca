package db

import (
	"github.com/bakape/meguca/common"
	r "github.com/dancannon/gorethink"
)

// Preconstructed REQL queries that don't have to be rebuilt
var (
	// Retrieves all threads for the /all/ metaboard
	getAllBoard = r.
			Table("threads").
			EqJoin("id", r.Table("posts")).
			Zip().
			Filter(filterDeleted).
			Without(omitForBoards).
			Merge(mergeLastUpdated).
			OrderBy(r.Desc("replyTime"))

	// Gets the most recently updated post timestamp from thread
	getLastUpdated = r.
			Table("posts").
			GetAllByIndex("op", r.Row.Field("id")).
			Field("lastUpdated").
			Max().
			Default(0)

	mergeLastUpdated = map[string]r.Term{
		"lastUpdated": getLastUpdated,
	}

	// Fields to omit in board queries. Decreases payload of DB replies.
	omitForBoards = []string{
		"body", "password", "commands", "links", "backlinks", "ip", "editing",
		"op",
	}

	// Fields to omit for post queries
	omitForPosts       = []string{"password", "ip", "lastUpdated"}
	omitForThreadPosts = append(omitForPosts, []string{"op", "board"}...)
)

func filterDeleted(p r.Term) r.Term {
	return p.Field("deleted").Default(false).Eq(false)
}

// GetThread retrieves public thread data from the database
func GetThread(id uint64, lastN int) (common.Thread, error) {
	q := r.
		Table("threads").
		GetAll(id). // Can not join after Get(). Meh.
		EqJoin("id", r.Table("posts")).
		Zip()

	getPosts := r.
		Table("posts").
		GetAllByIndex("op", id).
		Filter(filterDeleted).
		OrderBy("id").
		CoerceTo("array")

	// Only fetch last N number of replies
	if lastN != 0 {
		getPosts = getPosts.Slice(-lastN)
	}

	q = q.Merge(map[string]r.Term{
		"posts":       getPosts.Without(omitForThreadPosts),
		"lastUpdated": getLastUpdated,
	}).
		Without("ip", "op", "password")

	var thread common.Thread
	if err := One(q, &thread); err != nil {
		return thread, err
	}

	if thread.Deleted {
		return common.Thread{}, r.ErrEmptyResult
	}

	// Remove OP from posts slice to prevent possible duplication. Post might
	// be deleted before the thread due to a deletion race.
	if len(thread.Posts) != 0 && thread.Posts[0].ID == id {
		thread.Posts = thread.Posts[1:]
	}

	thread.Abbrev = lastN != 0

	return thread, nil
}

// GetPost reads a single post from the database
func GetPost(id uint64) (post common.StandalonePost, err error) {
	q := FindPost(id).Without(omitForPosts).Default(nil)
	err = One(q, &post)
	if post.Deleted && err == nil {
		return common.StandalonePost{}, r.ErrEmptyResult
	}
	return
}

// GetBoard retrieves all OPs of a single board
func GetBoard(board string) (data common.Board, err error) {
	q := r.
		Table("threads").
		GetAllByIndex("board", board).
		EqJoin("id", r.Table("posts")).
		Zip().
		Filter(filterDeleted).
		Without(omitForBoards).
		Merge(mergeLastUpdated).
		OrderBy(r.Desc("replyTime"))
	err = All(q, &data)

	// So that JSON always produces and array
	if data == nil {
		data = common.Board{}
	}

	return
}

// GetAllBoard retrieves all threads for the "/all/" meta-board
func GetAllBoard() (data common.Board, err error) {
	err = All(getAllBoard, &data)
	if data == nil {
		data = common.Board{}
	}
	return
}
