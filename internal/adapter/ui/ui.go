package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/rivo/tview"
	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/models"
)

type api interface {
	EventRegistration(login, password, password2 string) error
	EventLogout() error
	EventAuthorization(login, password string) error
	EventGetMetaDatas() (*[]models.FileMetaDataItem, error)
	EventNewCard(eID uint, title, number, cvv, pin, date string) error
	EventGetCard(id int64) (*models.Card, error)
	EventEditCard(id int64, title, number, cvv, pin, date string) error
	EventDeleteCard(id int64) error
	EventNewText(eID uint, title, text string) (*models.FileMetaDataItem, error)
	EventGetText(id int64) (*models.Text, error)
	EventEditText(id int64, title, text string) error
	EventDeleteText(id int64) error
	EventNewPassword(eID uint, title, site, login, password string) (*models.FileMetaDataItem, error)
	EventGetPassword(id int64) (*models.Password, error)
	EventEditPassword(id int64, title, site, login, password string) error
	EventDeletePassword(id int64) error
	EventNewFile(eID uint, title, path string) (*models.FileMetaDataItem, error)
	EventGetFile(id int64) (*models.Binary, error)
	EventEditFile(id int64, title, path string) error
	EventUploadFile(id int64, path string) error
	EventDeleteFile(id int64) error
}

var (
	btnLabelExit         = "Выйти"
	btnLabelAdd          = "Добавить"
	btnLabelSave         = "Сохранить"
	btnLabelDelete       = "Удалить"
	btnLableBack         = "Назад"
	inputLabelTitle      = "Название"
	inputLabelNumberCard = "Номер карты"
	inputLabelCVV        = "CVV"
	inputLabelPing       = "Pin код"

	maxLenCVV        = 3
	maxLenPIN        = 6
	maxLenNumberCard = 16
	maxLenPath       = 255

	errGetData = "Ошибка получения данных"
)

type terminal struct {
	app     *tview.Application
	api     api
	log     *zap.Logger
	version string
	date    string
	commit  string
}

type option func(*terminal)

func SetVersion(version, date, commit string) option {
	return func(t *terminal) {
		t.version = version
		t.date = date
		t.commit = commit
	}
}

// New создаем терминальный клиент.
func New(ctx context.Context, api api, log *zap.Logger, options ...option) (*terminal, error) {
	t := &terminal{
		app: tview.NewApplication(),
		api: api,
		log: log,
	}

	for _, opt := range options {
		opt(t)
	}

	return t, nil
}

// Run запускаем терминальный клиент.
func (t *terminal) Run(isDraw *bool) error {
	t.startPage(
		func() { t.authPage() },
		func() { t.registratingPage() },
		isDraw,
	)
	return nil
}

// Close закрываем терминал.
func (t *terminal) Close() {
	if err := t.api.EventLogout(); err != nil {
		t.log.Error("failed logout event", zap.Error(err))
	}
	t.app.Stop()
}

func (t *terminal) startPage(btnSignIn func(), btnReg func(), isDraw *bool) {
	lenVersionString := 20
	form := tview.NewForm().
		AddTextView("Версия", t.version, lenVersionString, 1, true, false).
		AddTextView("Коммит", t.commit, lenVersionString, 1, false, false).
		AddTextView("Дата", t.date, lenVersionString, 1, false, false).
		AddButton("Войти", btnSignIn).
		AddButton("Регистрация", btnReg).
		AddButton(btnLabelExit, t.Close)
	form.SetBorder(true).SetTitle("GophKeeper").SetTitleAlign(tview.AlignLeft)
	if isDraw != nil && *isDraw {
		t.app.SetRoot(form, true).SetFocus(form).EnableMouse(true).ForceDraw()
	} else {
		if err := t.app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
			t.errorPage(fmt.Sprintf("failed create start from: %v", err), btnSignIn)
		}
	}
}

func (t *terminal) authPage() {
	var login string
	var password string
	isDraw := true
	lenInput := 20

	form := tview.NewForm().
		AddInputField("Login", "", lenInput, nil, func(text string) { login = text }).
		AddPasswordField("Password", "", lenInput, '*', func(text string) { password = text }).
		AddButton("Войти", func() {
			err := t.api.EventAuthorization(login, password)
			if err != nil {
				t.errorPage(err.Error(), t.authPage)
				return
			}
			t.mainPage()
		}).
		AddButton(btnLableBack, func() {
			if err := t.Run(&isDraw); err != nil {
				t.errorPage(err.Error(), func() {
					t.authPage()
				})
			}
		}).
		AddButton(btnLabelExit, t.Close)
	form.SetBorder(true).SetTitle("Авторизация").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).EnableMouse(true).ForceDraw()
}

func (t *terminal) registratingPage() {
	var login string
	var password string
	var password2 string
	isDraw := true
	letField := 20

	form := tview.NewForm().
		AddInputField("Login", "", letField, nil, func(text string) { login = text }).
		AddPasswordField("Password", "", letField, '*', func(text string) { password = text }).
		AddPasswordField("Password2", "", letField, '*', func(text string) { password2 = text }).
		AddButton("Зарегистрироваться", func() {
			err := t.api.EventRegistration(login, password, password2)
			if err != nil {
				t.errorPage(fmt.Sprintf("Ошибка регистрации: %v", err), func() { t.registratingPage() })
				return
			}
			t.modal("Вы успешно зарегестрировались", map[string]func(){
				"Ok": func() {
					if err := t.Run(&isDraw); err != nil {
						t.errorPage(err.Error(), func() {
							t.authPage()
						})
					}
				},
			})
		}).
		AddButton(btnLableBack, func() {
			if err := t.Run(&isDraw); err != nil {
				t.errorPage(err.Error(), func() {
					t.authPage()
				})
			}
		}).
		AddButton(btnLabelExit, t.Close)
	form.SetBorder(true).SetTitle("Регистрация").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).EnableMouse(true).ForceDraw()
}

func (t *terminal) mainPage() {
	list := tview.NewList()

	data, err := t.api.EventGetMetaDatas()
	if err != nil {
		t.errorPage(err.Error(), func() { t.mainPage() })
		return
	}
	for i, e := range *data {
		if !e.IsDeleted {
			list = list.AddItem(fmt.Sprintf("%s | %s", string(e.DataType), e.Title), "", rune(i), func() {
				switch e.DataType {
				case models.CARD:
					t.editCardPage(e.ID)
				case models.TEXT:
					t.editTextPage(e.ID)
				case models.PASSWORD:
					t.editPasswordPage(e.ID)
				case models.BINARY:
					t.editFilePage(e.ID)
				}
			})
		}
	}

	list.
		AddItem("Добавить текст", "", 't', func() { t.newTextPage() }).
		AddItem("Добавить карту", "", 'c', func() { t.newCardPage() }).
		AddItem("Добавить пару логин/пароль", "", 'p', func() { t.newPasswordPage() }).
		AddItem("Добавить файл", "", 'f', func() { t.newFilePage() }).
		AddItem("Обновить", "", 'r', func() { t.mainPage() }).
		AddItem(btnLabelExit, "Press to exit", 'q', t.Close).
		SetBorder(true).SetTitle("Список сохраненных данных")

	t.app.SetRoot(list, true).SetFocus(list).EnableMouse(true).ForceDraw()
}

func (t *terminal) errorPage(message string, okBtn func()) {
	width := 100
	height := 5
	form := tview.NewForm().
		AddTextView("Error", message, width, height, false, false).
		AddButton("Ok", okBtn)
	form.SetBorder(true).SetTitle("Error").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).EnableMouse(true).ForceDraw()
}

func (t *terminal) newCardPage() {
	var title string
	var number string
	var cvv string
	var pin string
	var date string

	lenLabel := 20
	lenCode := 10
	lenDate := 11
	form := tview.NewForm().
		AddInputField(inputLabelTitle, "", lenLabel, nil, func(text string) { title = text }).
		AddInputField(inputLabelNumberCard, "", lenLabel, summCheck(isNumber(), length(maxLenNumberCard)),
			func(text string) { number = text }).
		AddInputField(inputLabelCVV, "", lenCode, summCheck(isNumber(), length(maxLenCVV)),
			func(text string) { cvv = text }).
		AddInputField(inputLabelPing, "", lenCode, summCheck(isNumber(), length(maxLenPIN)),
			func(text string) { pin = text }).
		AddInputField("Дата", time.Now().Format(time.DateOnly), lenDate, nil, func(text string) { date = text }).
		AddButton(btnLabelAdd, func() {
			err := t.api.EventNewCard(0, title, number, cvv, pin, date)
			if err != nil {
				t.errorPage(err.Error(), func() { t.newCardPage() })
				return
			}
			t.mainPage()
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Добавить новую карту").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) editCardPage(id int64) {
	card, err := t.api.EventGetCard(id)
	if err != nil {
		t.errorPage(errGetData, func() { t.mainPage() })
		return
	}
	lenLong := 20
	lenShort := 10

	form := tview.NewForm().
		AddInputField(inputLabelTitle, card.Title, lenLong, nil, func(text string) { card.Title = text }).
		AddInputField(inputLabelNumberCard, card.Number, lenLong, summCheck(isNumber(), length(maxLenNumberCard)),
			func(text string) { card.Number = text }).
		AddInputField(inputLabelCVV, card.CVV, lenShort, summCheck(isNumber(), length(maxLenCVV)),
			func(text string) { card.CVV = text }).
		AddInputField(inputLabelPing, card.PIN, lenShort, summCheck(isNumber(), length(maxLenPIN)),
			func(text string) { card.PIN = text }).
		AddInputField("Дата", card.Expiry, lenShort, nil, func(text string) { card.Expiry = text }).
		AddButton(btnLabelSave, func() {
			err := t.api.EventEditCard(id, card.Title, card.Number, card.CVV, card.PIN, card.Expiry)
			if err != nil {
				t.errorPage(err.Error(), func() {
					t.editCardPage(id)
				})
			}
			t.modal("Карта обновлена", map[string]func(){
				"Ok": func() {
					t.mainPage()
				},
			})
		}).
		AddButton(btnLabelDelete, func() {
			t.modal(fmt.Sprintf("Удалить карту `%s`", card.Title), map[string]func(){
				"Да": func() {
					err := t.api.EventDeleteCard(id)
					if err != nil {
						t.errorPage(
							fmt.Sprintf("Ошибка удаления карты `%v`: %v", id, err),
							func() { t.editCardPage(id) },
						)
						return
					}
					t.mainPage()
				},
				"Отмена": func() { t.editCardPage(id) },
			})
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Редактор карты").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) newTextPage() {
	var title string
	var text_ string
	lenLong := 20
	width := 10
	height := 1000
	form := tview.NewForm().
		AddInputField(inputLabelTitle, "", lenLong, nil, func(text string) { title = text }).
		AddTextArea("Текст", "", lenLong, width, height, func(text string) { text_ = text }).
		AddButton(btnLabelAdd, func() {
			_, err := t.api.EventNewText(0, title, text_)
			if err != nil {
				t.errorPage(err.Error(), func() { t.newTextPage() })
				return
			}
			t.mainPage()
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Добавить новую карту").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) editTextPage(id int64) {
	txt, err := t.api.EventGetText(id)
	if err != nil {
		t.errorPage(errGetData, func() { t.mainPage() })
		return
	}
	lenLong := 40
	width := 25
	height := 1000
	form := tview.NewForm().
		AddInputField(inputLabelTitle, txt.Title, lenLong, nil, func(text string) { txt.Title = text }).
		AddTextArea("Текст", txt.Text, lenLong, width, height, func(text string) { txt.Text = text }).
		AddButton(btnLabelSave, func() {
			err := t.api.EventEditText(id, txt.Title, txt.Text)
			if err != nil {
				t.errorPage(err.Error(), func() { t.newTextPage() })
				return
			}
			t.mainPage()
		}).
		AddButton(btnLabelDelete, func() {
			t.modal(fmt.Sprintf("Удалить текст `%s`", txt.Title), map[string]func(){
				"Да": func() {
					err := t.api.EventDeleteText(id)
					if err != nil {
						t.errorPage(fmt.Sprintf("Ошибка удаления текста `%v`: %v", id, err), func() { t.editTextPage(id) })
						return
					}
					t.mainPage()
				},
				"Отмена": func() { t.editCardPage(id) },
			})
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Изменить текст").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) newPasswordPage() {
	var title string
	var site string
	var login string
	var password string
	lenInput := 20
	form := tview.NewForm().
		AddInputField(inputLabelTitle, "", lenInput, nil, func(text string) { title = text }).
		AddInputField("Сайт", "", lenInput, nil, func(text string) { site = text }).
		AddInputField("Логин", "", lenInput, nil, func(text string) { login = text }).
		AddInputField("Пароль", "", lenInput, nil, func(text string) { password = text }).
		AddButton(btnLabelAdd, func() {
			_, err := t.api.EventNewPassword(0, title, site, login, password)
			if err != nil {
				t.errorPage(err.Error(), func() { t.newPasswordPage() })
				return
			}
			t.mainPage()
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Добавить новую карту").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) modal(message string, btns map[string]func()) {
	modal := tview.NewModal().
		SetText(message)
	labels := []string{}
	for k := range btns {
		labels = append(labels, k)
	}
	modal.AddButtons(labels).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			for label, f := range btns {
				if buttonLabel == label {
					f()
				}
			}
		})
	t.app.SetRoot(modal, false).SetFocus(modal).ForceDraw()
}

func (t *terminal) editPasswordPage(id int64) {
	psw, err := t.api.EventGetPassword(id)
	if err != nil {
		t.errorPage(errGetData, func() { t.mainPage() })
		return
	}
	lenLong := 40
	lenShort := 20
	form := tview.NewForm().
		AddInputField(inputLabelTitle, psw.Title, lenLong, nil, func(text string) { psw.Title = text }).
		AddInputField("Сайт", psw.Site, lenLong, nil, func(text string) { psw.Site = text }).
		AddInputField("Логин", psw.Login, lenShort, nil, func(text string) { psw.Login = text }).
		AddInputField("Пароль", psw.Password, lenShort, nil, func(text string) { psw.Password = text }).
		AddButton(btnLabelSave, func() {
			err := t.api.EventEditPassword(id, psw.Title, psw.Site, psw.Login, psw.Password)
			if err != nil {
				t.errorPage(err.Error(), func() { t.editPasswordPage(id) })
				return
			}
			t.mainPage()
		}).
		AddButton(btnLabelDelete, func() {
			t.modal(fmt.Sprintf("Удалить пароль `%s`", psw.Title), map[string]func(){
				"Да": func() {
					err := t.api.EventDeleteCard(id)
					if err != nil {
						t.errorPage(fmt.Sprintf("Ошибка удаления пароля `%v`: %v", id, err), func() { t.editPasswordPage(id) })
						return
					}
					t.mainPage()
				},
				"Отмена": func() { t.editPasswordPage(id) },
			})
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Обновить пароль").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) newFilePage() {
	var title string
	var path string
	lenTitle := 20
	lenPath := 100

	form := tview.NewForm().
		AddInputField(inputLabelTitle, "", lenTitle, nil, func(text string) { title = text }).
		AddInputField("Путь", "", lenPath, summCheck(length(maxLenPath)), func(text string) { path = text }).
		AddTextView("", "Укажите полный путь до файла", lenPath, 1, false, false).
		AddButton(btnLabelAdd, func() {
			_, err := t.api.EventNewFile(0, title, path)
			if err != nil {
				t.errorPage(err.Error(), func() { t.newFilePage() })
				return
			}
			t.mainPage()
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Добавить файл").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) editFilePage(id int64) {
	var path string
	file, err := t.api.EventGetFile(id)
	if err != nil {
		t.errorPage(errGetData, func() { t.mainPage() })
		return
	}
	lenTitle := 20
	lenPath := 100
	form := tview.NewForm().
		AddInputField(inputLabelTitle, file.Title, lenTitle, nil, func(text string) { file.Title = text }).
		AddInputField("Путь", "", lenPath, summCheck(length(maxLenPath)), func(text string) { path = text }).
		AddTextView("", "Укажите полный путь до файла", lenPath, 1, false, false).
		AddButton(btnLabelSave, func() {
			err := t.api.EventEditFile(id, file.Title, path)
			if err != nil {
				t.errorPage(err.Error(), func() {
					t.editFilePage(id)
				})
			}
			t.modal("Файл обновлен", map[string]func(){
				"Ok": func() {
					t.mainPage()
				},
			})
		}).
		AddButton("Скачать "+file.Filename, func() {
			t.uploadFilePage(id)
		}).
		AddButton(btnLabelDelete, func() {
			t.modal(fmt.Sprintf("Удалить файл `%s`", file.Title), map[string]func(){
				"Да": func() {
					err := t.api.EventDeleteFile(id)
					if err != nil {
						t.errorPage(fmt.Sprintf("Ошибка удаления файла `%v`: %v", id, err), func() { t.editFilePage(id) })
						return
					}
					t.mainPage()
				},
				"Отмена": func() { t.editFilePage(id) },
			})
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Обновление файла").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}

func (t *terminal) uploadFilePage(id int64) {
	file, err := t.api.EventGetFile(id)
	if err != nil {
		t.errorPage(errGetData, func() { t.mainPage() })
		return
	}
	var path string
	lenPath := 100
	form := tview.NewForm().
		AddInputField("Укажите директорию для скачивания", "", lenPath, nil, func(text string) { path = text }).
		AddButton("Скачать "+file.Filename, func() {
			err := t.api.EventUploadFile(id, path)
			if err != nil {
				t.errorPage(err.Error(), func() { t.uploadFilePage(id) })
				return
			}
			t.modal("Файл скачан "+path+"/"+file.Filename, map[string]func(){
				"Ok": func() { t.editFilePage(id) },
			})
		}).
		AddButton(btnLableBack, func() { t.mainPage() })
	form.SetBorder(true).SetTitle("Редактор карты").SetTitleAlign(tview.AlignLeft)
	t.app.SetRoot(form, true).SetFocus(form).ForceDraw()
}
