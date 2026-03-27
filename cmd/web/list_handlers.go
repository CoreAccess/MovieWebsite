package main

import (
	"filmgap/internal/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (app *application) userListsView(w http.ResponseWriter, r *http.Request) {
	user := app.getUser(r)
	if user == nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}
	
	lists, err := app.Service.GetListsByUserID(user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.getTemplateData("My Lists", r)
	data.Lists = lists

	app.render(w, r, http.StatusOK, "lists.html", data)
}

func (app *application) listView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	slug := r.PathValue("slug")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.notFound(w)
		return
	}

	// Fetch detail (metadata)
	list, err := app.Service.GetListByID(id)
	if err != nil {
		app.notFound(w)
		return
	}

	// Verify slug matches for SEO
	if list.Slug != slug {
		http.Redirect(w, r, fmt.Sprintf("/list/%d/%s", list.ID, list.Slug), http.StatusMovedPermanently)
		return
	}

	// Fetch items
	listItems, err := app.Service.Repo.GetListItems(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Fetch follow counts
	followers, _, err := app.Service.GetListFollowCounts(id)
	if err == nil {
		list.FollowerCount = followers
	}

	data := app.getTemplateData(list.Name, r)
	data.List = list
	data.ListItems = listItems

	// If authenticated, check follow status
	currUser := app.getUser(r)
	if currUser != nil {
		isFollowing, _ := app.Service.IsFollowingList(currUser.ID, id)
		data.IsFollowing = isFollowing
	}

	app.render(w, r, http.StatusOK, "list_view.html", data)
}

func (app *application) listCreateView(w http.ResponseWriter, r *http.Request) {
	data := app.getTemplateData("Create New List", r)
	app.render(w, r, http.StatusOK, "list_create.html", data)
}

func (app *application) listCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	name := r.PostForm.Get("name")
	description := r.PostForm.Get("description")
	visibility := r.PostForm.Get("visibility")
	ranked := r.PostForm.Get("ranked") == "true"

	if name == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.getUser(r)
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-")) // Basic slugification

	list := models.List{
		UserID:      user.ID,
		Name:        name,
		Slug:        slug,
		Description: description,
		Visibility:  visibility,
		IsRanked:    ranked,
	}

	id, err := app.Service.CreateList(list)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/list/%d/%s", id, slug), http.StatusSeeOther)
}

func (app *application) listItemAddPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	listID, _ := strconv.Atoi(r.PostForm.Get("list_id"))
	mediaID, _ := strconv.Atoi(r.PostForm.Get("media_id"))
	note := r.PostForm.Get("note")
	rank, _ := strconv.Atoi(r.PostForm.Get("rank"))

	if listID == 0 || mediaID == 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.getUser(r)
	
	// Check ownership
	list, err := app.Service.GetListByID(listID)
	if err != nil {
		app.notFound(w)
		return
	}
	if list.UserID != user.ID {
		app.clientError(w, http.StatusForbidden)
		return
	}

	item := models.ListItem{
		ListID:  listID,
		MediaID: mediaID,
		Note:    note,
		Rank:    rank,
		AddedBy: user.ID,
	}

	err = app.Service.AddListItem(item)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect back to referring page or list view
	referer := r.Header.Get("Referer")
	if referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/list/%d/%s", listID, list.Slug), http.StatusSeeOther)
	}
}

func (app *application) listItemRemovePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	listID, _ := strconv.Atoi(r.PostForm.Get("list_id"))
	mediaID, _ := strconv.Atoi(r.PostForm.Get("media_id"))

	user := app.getUser(r)

	// Check ownership
	list, err := app.Service.GetListByID(listID)
	if err != nil {
		app.notFound(w)
		return
	}
	if list.UserID != user.ID {
		app.clientError(w, http.StatusForbidden)
		return
	}

	err = app.Service.RemoveListItem(listID, mediaID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/list/%d/%s", listID, list.Slug), http.StatusSeeOther)
}
