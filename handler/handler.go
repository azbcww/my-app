package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db *sqlx.DB
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

type LoginRequestBody struct {
	Username string `json:"username,omitempty" form:"username"`
	Password string `json:"password,omitempty" form:"password"`
}

type User struct {
	Username   string `json:"username,omitempty"  db:"Username"`
	HashedPass string `json:"-"  db:"HashedPass"`
}

type Me struct {
	Username string `json:"username,omitempty"  db:"username"`
}

type UserAndData struct {
	Username   string `json:"username,omitempty"  db:"Username"`
	Data   string `json:"data,omitempty"`
}

type DateData struct {
	Start string `json:"start"`
	End string `json:"end"`
	Error string `json:"error"`
}

type RemoveResponseData struct {
	Error string `json:"error"`
}

func (h *Handler) SignUpHandler(c echo.Context) error {
	// リクエストを受け取り、reqに格納する
	req := LoginRequestBody{}
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	// バリデーションする(PasswordかUsernameが空文字列の場合は400 BadRequestを返す)
	if req.Password == "" || req.Username == "" {
		return c.String(http.StatusBadRequest, "Username or Password is empty")
	}

	// 登録しようとしているユーザーが既にデータベース内に存在するかチェック
	var count int
	err = h.db.Get(&count, "SELECT COUNT(*) FROM users WHERE Username=?", req.Username)
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	// 存在したら409 Conflictを返す
	if count > 0 {
		return c.String(http.StatusConflict, "Username is already used")
	}

	// パスワードをハッシュ化する
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	// ハッシュ化に失敗したら500 InternalServerErrorを返す
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// ユーザーを登録する
	_, err = h.db.Exec("INSERT INTO users (Username, HashedPass) VALUES (?, ?)", req.Username, hashedPass)
	// 登録に失敗したら500 InternalServerErrorを返す
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// セッションストアに登録する
	sess, err := session.Get("sessions", c)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "something wrong in getting session")
	}
	sess.Values["userName"] = req.Username
	sess.Save(c.Request(), c.Response())
	
	// 登録に成功したら201 Createdを返す
	return c.NoContent(http.StatusCreated)
}

func (h *Handler) LoginHandler(c echo.Context) error {
	// リクエストを受け取り、reqに格納する
	var req LoginRequestBody
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request body")
	}

	// バリデーションする(PasswordかUsernameが空文字列の場合は400 BadRequestを返す)
	if req.Password == "" || req.Username == "" {
		return c.String(http.StatusBadRequest, "Username or Password is empty")
	}

	// データベースからユーザーを取得する
	user := User{}
	err = h.db.Get(&user, "SELECT * FROM users WHERE username=?", req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.NoContent(http.StatusUnauthorized)
		} else {
			log.Println(err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	// パスワードが一致しているかを確かめる
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPass), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return c.NoContent(http.StatusUnauthorized)
		} else {
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	// セッションストアに登録する
	sess, err := session.Get("sessions", c)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "something wrong in getting session")
	}
	sess.Values["userName"] = req.Username
	sess.Save(c.Request(), c.Response())

	return c.NoContent(http.StatusOK)
}

func UserAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, "something wrong in getting session")
		}
		if sess.Values["userName"] == nil {
			return c.String(http.StatusUnauthorized, "please login")
		}
		c.Set("userName", sess.Values["userName"].(string))
		return next(c)
	}
}

func GetMeHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, Me{
		Username: c.Get("userName").(string),
	})
}

func (h *Handler) RegisterEvent(c echo.Context) error {
	var data UserAndData
	err := c.Bind(&data)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	cmd := exec.Command("python3", "./main.py", data.Data)
	out, err := cmd.Output()

	if err != nil {
		log.Printf("failed to exec the python script: %s\n", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	var dateData DateData
	if err = json.Unmarshal(out, &dateData); err != nil {
		log.Printf("failed to unmarshal: %s\n", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// イベントを登録する
	_, err = h.db.Exec("INSERT INTO content (userName, startDate, endDate, title) VALUES (?, ?, ?, ?)", data.Username, dateData.Start, dateData.End, data.Data)
	// 登録に失敗したら500 InternalServerErrorを返す
	if err != nil {
		log.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, dateData)
}

func (h *Handler) RemoveEvent(c echo.Context) error {
	var data UserAndData
	err := c.Bind(&data)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	cmd := exec.Command("python3", "./main.py", data.Data)
	out, err := cmd.Output()

	if err != nil {
		log.Printf("failed to exec the python script: %s\n", err)
		var res RemoveResponseData
		res.Error = "11111"
		return c.JSON(http.StatusInternalServerError, res)
		// return c.NoContent(http.StatusInternalServerError)
	}
	var dateData DateData
	if err = json.Unmarshal(out, &dateData); err != nil {
		log.Printf("failed to unmarshal: %s\n", err)
		var res RemoveResponseData
		res.Error = "222222"
		return c.JSON(http.StatusInternalServerError, res)
		// return c.NoContent(http.StatusInternalServerError)
	}
	// イベントを削除する
	_, err = h.db.Exec("DELETE FROM content  WHERE userName=? AND startDate=? AND endDate=? AND title=? LIMIT 1", data.Username, dateData.Start, dateData.End, data.Data)
	// 登録に失敗したら500 InternalServerErrorを返す
	if err != nil {
		log.Println(err)
		var res RemoveResponseData
		res.Error = err.Error()
		return c.JSON(http.StatusInternalServerError, res)
	}
	var res RemoveResponseData
	res.Error = ""
	return c.JSON(http.StatusCreated, res)
}