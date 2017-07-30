package actions

import (
	"fmt"
	"os"
	"time"

	"github.com/bscott/golangflow/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (Post)
// DB Table: Plural (Posts)
// Resource: Plural (Posts)
// Path: Plural (/posts)
// View Template Folder: Plural (/templates/posts/)

func init() {
	w := App().Worker
	w.Register("send_tweet", func(args worker.Args) error {
		fmt.Printf("### args -> %+v\n", args)
		shortURL, err := getBitly(args["post_id"])
		if err != nil {
			fmt.Errorf("Tweet Worker encountered an error with Bitly: %v", err)
		}
		return nil
	})
}

// PostsResource is the resource for the post model
type PostsResource struct {
	buffalo.Resource
}

// List gets all Posts. This function is mapped to the path
// GET /posts
func (v PostsResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)

	posts := &models.Posts{}
	// You can order your list here. Just change
	errp := tx.Where("user_id = ?", c.Value("current_user_id")).All(posts)

	// to:
	// err := tx.Order("create_at desc").All(posts)
	if errp != nil {
		return errors.WithStack(errp)
	}
	// Make posts available inside the html template
	c.Set("posts", posts)

	return c.Render(200, r.HTML("posts/index.html"))
}

// Show gets the data for one Post. This function is mapped to
// the path GET /posts/{post_id}
func (v PostsResource) Show(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty Post
	post := &models.Post{}

	// To find the Post the parameter post_id is used.
	err := tx.Find(post, c.Param("post_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	// Make post available inside the html template

	c.Set("post", post)
	return c.Render(200, r.HTML("posts/show.html"))
}

// New renders the formular for creating a new post.
// This function is mapped to the path GET /posts/new
func (v PostsResource) New(c buffalo.Context) error {
	// Make post available inside the html template
	c.Set("post", &models.Post{})
	return c.Render(200, r.HTML("posts/new.html"))
}

// Create adds a post to the DB. This function is mapped to the
// path POST /posts
func (v PostsResource) Create(c buffalo.Context) error {
	// Search for current logged in user
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty User
	// Allocate an empty Post
	post := &models.Post{UserID: c.Value("current_user_id").(uuid.UUID)}
	// Bind post to the html form elements
	err := c.Bind(post)

	if err != nil {
		return errors.WithStack(err)
	}

	// Get the DB connection from the context
	//tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(post)

	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		// Make post available inside the html template
		c.Set("post", post)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		// Render again the new.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("posts/new.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "Post was created successfully")

	// Queue tweet
	w := App().Worker
	w.PerformIn(worker.Job{
		Queue: "tweet",
		Args: worker.Args{
			"post_id":      post.ID,
			"post_content": post.Title,
		},
		Handler: "send_tweet",
	}, 15*time.Second)

	// and redirect to the posts index page
	return c.Redirect(302, "/posts/%s", post.ID)
}

// Edit renders a edit formular for a post. This function is
// mapped to the path GET /posts/{post_id}/edit
func (v PostsResource) Edit(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty Post
	post := &models.Post{}
	err := tx.Where("user_id = ?", c.Value("current_user_id")).Find(post, c.Param("post_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	// Make post available inside the html template
	c.Set("post", post)
	return c.Render(200, r.HTML("posts/edit.html"))
}

// Update changes a post in the DB. This function is mapped to
// the path PUT /posts/{post_id}
func (v PostsResource) Update(c buffalo.Context) error {

	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty Post
	post := &models.Post{}
	err := tx.Where("user_id = ?", c.Value("current_user_id")).Find(post, c.Param("post_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	// Bind post to the html form elements
	err = c.Bind(post)
	if err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(post)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		// Make post available inside the html template
		c.Set("post", post)
		// Make the errors available inside the html template
		c.Set("errors", verrs)
		// Render again the edit.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("posts/edit.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "Post was updated successfully")
	// and redirect to the posts index page
	return c.Redirect(302, "/posts/%s", post.ID)
}

// Destroy deletes a post from the DB. This function is mapped
// to the path DELETE /posts/{post_id}
func (v PostsResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Allocate an empty Post
	post := &models.Post{}
	// To find the Post the parameter post_id is used.
	err := tx.Where("user_id = ?", c.Value("current_user_id")).Find(post, c.Param("post_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = tx.Destroy(post)
	if err != nil {
		return errors.WithStack(err)
	}
	// If there are no errors set a flash message
	c.Flash().Add("success", "Post was destroyed successfully")
	// Redirect to the posts index page
	return c.Redirect(302, "/posts")
}

// Tweet functions

func getBitly(id uuid.UUID) (bitly.ShortenResult, error) {
	// Load Bitly config data
	accessToken := os.Getenv("BITLY_ACCESS_TOKEN")
	bitlyLogin := os.Getenv("BITLY_LOGIN")
	bitlyAPIKey := os.Getenv("BITLY_API_KEY")

	url := "http://golangflow.io/posts/" + string(id)

	c := bitly.Client{AccessToken: accessToken, Login: bitlyLogin, APIKey: bitlyAPIKey}
	c.s

	return "", nil
}
