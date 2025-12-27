package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"safe-wallet-go/pkg"
)

// VaultApp is the main application structure
type VaultApp struct {
	app         fyne.App
	mainWindow  fyne.Window
	service     *pkg.WalletService
	filepath    string
	currentPath pkg.Path

	// UI Components
	treeWidget   *widget.Tree
	detailsPanel *fyne.Container
	searchEntry  *widget.Entry
	statusLabel  *widget.Label
	breadcrumbs  *widget.Label
}

func NewVaultApp() *VaultApp {
	myApp := app.NewWithID("com.safewallet.go")
	myApp.Settings().SetTheme(theme.DarkTheme())

	return &VaultApp{
		app:         myApp,
		filepath:    "wallet.dat",
		currentPath: pkg.Path{GroupIDs: []string{}},
	}
}

func (va *VaultApp) Run() {
	va.mainWindow = va.app.NewWindow("Safe Wallet - Password Vault")
	va.mainWindow.Resize(fyne.NewSize(1200, 700))
	va.mainWindow.CenterOnScreen()

	// Show unlock screen first
	va.showUnlockScreen()

	va.mainWindow.ShowAndRun()
}

func (va *VaultApp) showUnlockScreen() {
	title := widget.NewLabel("Safe Wallet")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	var content *fyne.Container

	if !pkg.WalletExists(va.filepath) {
		// Create new wallet
		subtitle := widget.NewLabel("Create a new vault")
		subtitle.Alignment = fyne.TextAlignCenter

		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.SetPlaceHolder("Master Password")

		confirmEntry := widget.NewPasswordEntry()
		confirmEntry.SetPlaceHolder("Confirm Password")

		createBtn := widget.NewButton("Create Vault", func() {
			if passwordEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("password cannot be empty"), va.mainWindow)
				return
			}

			if passwordEntry.Text != confirmEntry.Text {
				dialog.ShowError(fmt.Errorf("passwords do not match"), va.mainWindow)
				return
			}

			va.service = pkg.NewWalletService(va.filepath, passwordEntry.Text)
			if err := va.service.CreateNew(); err != nil {
				dialog.ShowError(fmt.Errorf("failed to create wallet: %v", err), va.mainWindow)
				return
			}

			dialog.ShowInformation("Success", "Wallet created successfully!", va.mainWindow)
			va.showMainInterface()
		})
		createBtn.Importance = widget.HighImportance

		content = container.NewVBox(
			layout.NewSpacer(),
			container.NewCenter(
				container.NewVBox(
					title,
					subtitle,
					widget.NewSeparator(),
					passwordEntry,
					confirmEntry,
					createBtn,
				),
			),
			layout.NewSpacer(),
		)
	} else {
		// Unlock existing wallet
		subtitle := widget.NewLabel("Enter your master password to unlock")
		subtitle.Alignment = fyne.TextAlignCenter

		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.SetPlaceHolder("Master Password")

		unlockBtn := widget.NewButton("Unlock Vault", func() {
			if passwordEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("password cannot be empty"), va.mainWindow)
				return
			}

			va.service = pkg.NewWalletService(va.filepath, passwordEntry.Text)
			if err := va.service.Load(); err != nil {
				dialog.ShowError(fmt.Errorf("failed to load wallet: %v", err), va.mainWindow)
				return
			}

			va.showMainInterface()
		})
		unlockBtn.Importance = widget.HighImportance

		content = container.NewVBox(
			layout.NewSpacer(),
			container.NewCenter(
				container.NewVBox(
					title,
					subtitle,
					widget.NewSeparator(),
					passwordEntry,
					unlockBtn,
				),
			),
			layout.NewSpacer(),
		)
	}

	va.mainWindow.SetContent(content)
}

func (va *VaultApp) showMainInterface() {
	// Create toolbar
	toolbar := va.createToolbar()

	// Create breadcrumbs
	va.breadcrumbs = widget.NewLabel(va.getBreadcrumbText())
	va.breadcrumbs.TextStyle = fyne.TextStyle{Bold: true}

	// Create search bar
	va.searchEntry = widget.NewEntry()
	va.searchEntry.SetPlaceHolder("Search entries...")
	va.searchEntry.OnChanged = func(s string) {
		if s != "" {
			va.performSearch(s)
		}
	}

	searchContainer := container.NewBorder(nil, nil,
		widget.NewIcon(theme.SearchIcon()), nil, va.searchEntry)

	// Create tree widget for groups
	va.treeWidget = va.createTreeWidget()

	// Create initial details panel
	va.detailsPanel = container.NewCenter(
		widget.NewLabel("Select a group or entry to view details"),
	)

	// Create status bar
	va.statusLabel = widget.NewLabel(va.getStatusText())
	statusBar := container.NewBorder(nil, nil, nil, nil, va.statusLabel)

	// Layout
	leftPanel := container.NewBorder(
		container.NewVBox(va.breadcrumbs, searchContainer),
		nil, nil, nil,
		container.NewScroll(va.treeWidget),
	)

	split := container.NewHSplit(
		leftPanel,
		va.detailsPanel,
	)
	split.SetOffset(0.35)

	content := container.NewBorder(
		toolbar,
		statusBar,
		nil, nil,
		split,
	)

	va.mainWindow.SetContent(content)
}

func (va *VaultApp) createToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderNewIcon(), func() {
			va.showAddGroupDialog()
		}),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			va.showAddEntryDialog()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			va.saveVault()
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			va.refreshTree()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			va.showSearchDialog()
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			va.currentPath = pkg.Path{GroupIDs: []string{}}
			va.refreshUI()
		}),
		widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
			va.navigateBack()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.LogoutIcon(), func() {
			va.lockVault()
		}),
	)
}

func (va *VaultApp) createTreeWidget() *widget.Tree {
	tree := widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return va.getChildUIDs(uid)
		},
		func(uid widget.TreeNodeID) bool {
			return va.isBranch(uid)
		},
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.FolderIcon())
			label := widget.NewLabel("Template")
			return container.NewHBox(icon, label)
		},
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			icon := c.Objects[0].(*widget.Icon)
			label := c.Objects[1].(*widget.Label)

			if uid == "" {
				icon.SetResource(theme.HomeIcon())
				label.SetText("ROOT")
			} else if branch {
				// It's a group
				parts := strings.Split(uid, "|")
				path := va.parseTreePath(parts)
				group, err := pkg.FindGroupByPath(va.service.GetWallet(), path)
				if err == nil {
					icon.SetResource(theme.FolderIcon())
					label.SetText(fmt.Sprintf("%s (%d)", group.Name, len(group.Entries)))
				}
			} else {
				// It's an entry
				parts := strings.Split(uid, "|")
				if len(parts) > 0 {
					entryID := parts[len(parts)-1]
					if strings.HasPrefix(entryID, "E:") {
						entryID = strings.TrimPrefix(entryID, "E:")
						groupPath := va.parseTreePath(parts[:len(parts)-1])
						group, err := pkg.FindGroupByPath(va.service.GetWallet(), groupPath)
						if err == nil {
							for _, entry := range group.Entries {
								if entry.ID == entryID {
									icon.SetResource(theme.DocumentIcon())
									label.SetText(entry.Title)
									break
								}
							}
						}
					}
				}
			}
		},
	)

	tree.OnSelected = func(uid widget.TreeNodeID) {
		if uid == "" {
			va.currentPath = pkg.Path{GroupIDs: []string{}}
			va.showGroupDetails(nil)
		} else {
			parts := strings.Split(uid, "|")
			lastPart := parts[len(parts)-1]

			if strings.HasPrefix(lastPart, "E:") {
				// It's an entry
				va.showEntryFromTree(parts)
			} else {
				// It's a group
				path := va.parseTreePath(parts)
				va.currentPath = path
				group, err := pkg.FindGroupByPath(va.service.GetWallet(), path)
				if err == nil {
					va.showGroupDetails(group)
				}
			}
		}
		va.updateBreadcrumbs()
	}

	return tree
}

func (va *VaultApp) getChildUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	wallet := va.service.GetWallet()
	if wallet == nil {
		return []widget.TreeNodeID{}
	}

	var children []widget.TreeNodeID

	if uid == "" {
		// Root level - return root groups
		for _, group := range wallet.Groups {
			children = append(children, group.ID)
		}
	} else {
		parts := strings.Split(uid, "|")
		lastPart := parts[len(parts)-1]

		// Check if this is an entry (entries have no children)
		if strings.HasPrefix(lastPart, "E:") {
			return []widget.TreeNodeID{}
		}

		// It's a group - get its children
		path := va.parseTreePath(parts)
		group, err := pkg.FindGroupByPath(wallet, path)
		if err != nil {
			return []widget.TreeNodeID{}
		}

		// Add subgroups
		for _, subgroup := range group.Groups {
			childUID := uid + "|" + subgroup.ID
			children = append(children, childUID)
		}

		// Add entries
		for _, entry := range group.Entries {
			childUID := uid + "|E:" + entry.ID
			children = append(children, childUID)
		}
	}

	return children
}

func (va *VaultApp) isBranch(uid widget.TreeNodeID) bool {
	if uid == "" {
		return true
	}

	parts := strings.Split(uid, "|")
	lastPart := parts[len(parts)-1]

	// Entries start with "E:"
	if strings.HasPrefix(lastPart, "E:") {
		return false
	}

	// It's a group
	return true
}

func (va *VaultApp) parseTreePath(parts []string) pkg.Path {
	var groupIDs []string
	for _, part := range parts {
		if !strings.HasPrefix(part, "E:") && part != "" {
			groupIDs = append(groupIDs, part)
		}
	}
	return pkg.Path{GroupIDs: groupIDs}
}

func (va *VaultApp) showEntryFromTree(parts []string) {
	entryID := strings.TrimPrefix(parts[len(parts)-1], "E:")
	groupPath := va.parseTreePath(parts[:len(parts)-1])

	group, err := pkg.FindGroupByPath(va.service.GetWallet(), groupPath)
	if err != nil {
		dialog.ShowError(err, va.mainWindow)
		return
	}

	for _, entry := range group.Entries {
		if entry.ID == entryID {
			va.showEntryDetails(entry, groupPath)
			return
		}
	}
}

func (va *VaultApp) showGroupDetails(group *pkg.Group) {
	var groupName string
	var subgroupCount, entryCount int

	if group == nil {
		groupName = "ROOT"
		subgroupCount = len(va.service.GetWallet().Groups)
		entryCount = 0
	} else {
		groupName = group.Name
		subgroupCount = len(group.Groups)
		entryCount = len(group.Entries)
	}

	title := widget.NewLabelWithStyle(groupName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	info := widget.NewLabel(fmt.Sprintf("Subgroups: %d | Entries: %d", subgroupCount, entryCount))

	buttons := container.NewHBox()
	if group != nil {
		editBtn := widget.NewButtonWithIcon("Rename", theme.DocumentCreateIcon(), func() {
			va.showEditGroupDialog(*group)
		})
		deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
			va.confirmDeleteGroup()
		})
		deleteBtn.Importance = widget.DangerImportance
		buttons = container.NewHBox(editBtn, deleteBtn)
	}

	details := container.NewVBox(
		title,
		widget.NewSeparator(),
		info,
		layout.NewSpacer(),
		buttons,
	)

	va.detailsPanel.Objects = []fyne.CanvasObject{details}
	va.detailsPanel.Refresh()
}

func (va *VaultApp) showEntryDetails(entry pkg.Entry, groupPath pkg.Path) {
	title := widget.NewLabelWithStyle(entry.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	fieldsContainer := container.NewVBox()

	for _, field := range entry.Fields {
		fieldLabel := widget.NewLabel(field.Name + ":")

		var valueWidget fyne.CanvasObject

		if field.Type == pkg.FieldTypePassword || field.Type == pkg.FieldTypePIN {
			passwordEntry := widget.NewPasswordEntry()
			passwordEntry.SetText(field.Value)
			passwordEntry.Disable()

			showBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), func(pe *widget.Entry) func() {
				return func() {
					if pe.Password {
						pe.Password = false
					} else {
						pe.Password = true
					}
					pe.Refresh()
				}
			}(passwordEntry))

			copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func(val string) func() {
				return func() {
					va.mainWindow.Clipboard().SetContent(val)
					dialog.ShowInformation("Copied", field.Name+" copied to clipboard", va.mainWindow)
				}
			}(field.Value))

			valueWidget = container.NewBorder(nil, nil, nil,
				container.NewHBox(showBtn, copyBtn), passwordEntry)
		} else {
			valueEntry := widget.NewEntry()
			valueEntry.SetText(field.Value)
			valueEntry.Disable()

			copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func(val string) func() {
				return func() {
					va.mainWindow.Clipboard().SetContent(val)
					dialog.ShowInformation("Copied", field.Name+" copied to clipboard", va.mainWindow)
				}
			}(field.Value))

			valueWidget = container.NewBorder(nil, nil, nil, copyBtn, valueEntry)
		}

		fieldsContainer.Add(fieldLabel)
		fieldsContainer.Add(valueWidget)
	}

	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		va.showEditEntryDialog(entry, groupPath)
	})

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		va.confirmDeleteEntry(entry, groupPath)
	})
	deleteBtn.Importance = widget.DangerImportance

	buttons := container.NewHBox(editBtn, deleteBtn)

	details := container.NewVBox(
		title,
		widget.NewSeparator(),
		fieldsContainer,
		layout.NewSpacer(),
		buttons,
	)

	scroll := container.NewScroll(details)
	va.detailsPanel.Objects = []fyne.CanvasObject{scroll}
	va.detailsPanel.Refresh()
}

func (va *VaultApp) showAddGroupDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Group Name")

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Name*", nameEntry),
		},
		OnSubmit: func() {
			if nameEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("group name cannot be empty"), va.mainWindow)
				return
			}

			group := &pkg.Group{
				Name:    nameEntry.Text,
				Groups:  []pkg.Group{},
				Entries: []pkg.Entry{},
			}

			if err := va.service.AddGroup(va.currentPath, group); err != nil {
				dialog.ShowError(fmt.Errorf("error creating group: %v", err), va.mainWindow)
				return
			}

			if err := va.service.Save(); err != nil {
				dialog.ShowError(fmt.Errorf("error saving: %v", err), va.mainWindow)
				return
			}

			dialog.ShowInformation("Success", "Group created successfully!", va.mainWindow)
			va.refreshTree()
		},
		OnCancel: func() {},
	}

	d := dialog.NewCustom("Add New Group", "Close", form, va.mainWindow)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}

func (va *VaultApp) showAddEntryDialog() {
	if len(va.currentPath.GroupIDs) == 0 {
		dialog.ShowError(fmt.Errorf("please select a group first"), va.mainWindow)
		return
	}

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("e.g., Gmail Account")

	// Template selection
	templates := []string{"Custom"}
	for _, t := range pkg.EntryTemplates {
		templates = append(templates, t.Name)
	}

	templateSelect := widget.NewSelect(templates, nil)
	templateSelect.SetSelected("Custom")

	fieldsContainer := container.NewVBox()
	var fields []pkg.EntryField

	updateFieldsUI := func(templateName string) {
		fieldsContainer.Objects = nil
		fields = []pkg.EntryField{}

		if templateName == "Custom" {
			// Show add field button for custom
			addFieldBtn := widget.NewButton("Add Field", func() {
				va.showAddFieldDialog(&fields, fieldsContainer)
			})
			fieldsContainer.Add(addFieldBtn)
		} else {
			// Show template fields
			for _, template := range pkg.EntryTemplates {
				if template.Name == templateName {
					for _, templateField := range template.Fields {
						field := pkg.EntryField{
							Name: templateField.Name,
							Type: templateField.Type,
						}
						fields = append(fields, field)

						label := widget.NewLabel(templateField.Name + "*:")
						var entry *widget.Entry
						if templateField.Type == pkg.FieldTypePassword || templateField.Type == pkg.FieldTypePIN {
							entry = widget.NewPasswordEntry()
						} else {
							entry = widget.NewEntry()
						}

						idx := len(fields) - 1
						entry.OnChanged = func(idx int) func(string) {
							return func(s string) {
								fields[idx].Value = s
							}
						}(idx)

						fieldsContainer.Add(label)
						fieldsContainer.Add(entry)
					}
					break
				}
			}
		}
		fieldsContainer.Refresh()
	}

	templateSelect.OnChanged = updateFieldsUI
	updateFieldsUI("Custom")

	scrollFields := container.NewScroll(fieldsContainer)
	scrollFields.SetMinSize(fyne.NewSize(400, 300))

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Title*:"),
			titleEntry,
			widget.NewLabel("Template:"),
			templateSelect,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		scrollFields,
	)

	d := dialog.NewCustomConfirm("Add New Entry", "Create", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}

		if titleEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("title cannot be empty"), va.mainWindow)
			return
		}

		if len(fields) == 0 {
			dialog.ShowError(fmt.Errorf("please add at least one field"), va.mainWindow)
			return
		}

		entry := &pkg.Entry{
			Title:  titleEntry.Text,
			Fields: fields,
		}

		if err := va.service.AddEntry(va.currentPath, entry); err != nil {
			dialog.ShowError(fmt.Errorf("error creating entry: %v", err), va.mainWindow)
			return
		}

		if err := va.service.Save(); err != nil {
			dialog.ShowError(fmt.Errorf("error saving: %v", err), va.mainWindow)
			return
		}

		dialog.ShowInformation("Success", "Entry created successfully!", va.mainWindow)
		va.refreshTree()
	}, va.mainWindow)

	d.Resize(fyne.NewSize(500, 600))
	d.Show()
}

func (va *VaultApp) showAddFieldDialog(fields *[]pkg.EntryField, fieldsContainer *fyne.Container) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Field Name")

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Field Value")

	typeSelect := widget.NewSelect([]string{"General", "Password", "PIN"}, nil)
	typeSelect.SetSelected("General")

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Name*", nameEntry),
			widget.NewFormItem("Value*", valueEntry),
			widget.NewFormItem("Type", typeSelect),
		},
		OnSubmit: func() {
			if nameEntry.Text == "" || valueEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("name and value cannot be empty"), va.mainWindow)
				return
			}

			fieldType := pkg.FieldTypeGeneral
			switch typeSelect.Selected {
			case "Password":
				fieldType = pkg.FieldTypePassword
			case "PIN":
				fieldType = pkg.FieldTypePIN
			}

			*fields = append(*fields, pkg.EntryField{
				Name:  nameEntry.Text,
				Value: valueEntry.Text,
				Type:  fieldType,
			})

			// Update UI
			label := widget.NewLabel(nameEntry.Text + ":")
			value := widget.NewLabel(valueEntry.Text)
			fieldsContainer.Add(container.NewHBox(label, value))
			fieldsContainer.Refresh()
		},
		OnCancel: func() {},
	}

	d := dialog.NewCustom("Add Field", "Close", form, va.mainWindow)
	d.Resize(fyne.NewSize(400, 250))
	d.Show()
}

func (va *VaultApp) showEditGroupDialog(group pkg.Group) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(group.Name)

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Name*", nameEntry),
		},
		OnSubmit: func() {
			if nameEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("group name cannot be empty"), va.mainWindow)
				return
			}

			updatedGroup := pkg.Group{
				Name:    nameEntry.Text,
				Groups:  group.Groups,
				Entries: group.Entries,
			}

			if err := va.service.UpdateGroup(va.currentPath, updatedGroup); err != nil {
				dialog.ShowError(fmt.Errorf("error updating group: %v", err), va.mainWindow)
				return
			}

			if err := va.service.Save(); err != nil {
				dialog.ShowError(fmt.Errorf("error saving: %v", err), va.mainWindow)
				return
			}

			dialog.ShowInformation("Success", "Group updated successfully!", va.mainWindow)
			va.refreshTree()
		},
		OnCancel: func() {},
	}

	d := dialog.NewCustom("Edit Group", "Close", form, va.mainWindow)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}

func (va *VaultApp) showEditEntryDialog(entry pkg.Entry, groupPath pkg.Path) {
	titleEntry := widget.NewEntry()
	titleEntry.SetText(entry.Title)

	fieldsContainer := container.NewVBox()
	editedFields := make([]pkg.EntryField, len(entry.Fields))
	copy(editedFields, entry.Fields)

	for i, field := range editedFields {
		idx := i
		fieldNameEntry := widget.NewEntry()
		fieldNameEntry.SetText(field.Name)
		fieldNameEntry.OnChanged = func(s string) {
			editedFields[idx].Name = s
		}

		var fieldValueEntry *widget.Entry
		if field.Type == pkg.FieldTypePassword || field.Type == pkg.FieldTypePIN {
			fieldValueEntry = widget.NewPasswordEntry()
		} else {
			fieldValueEntry = widget.NewEntry()
		}
		fieldValueEntry.SetText(field.Value)
		fieldValueEntry.OnChanged = func(s string) {
			editedFields[idx].Value = s
		}

		typeSelect := widget.NewSelect([]string{"General", "Password", "PIN"}, func(s string) {
			switch s {
			case "Password":
				editedFields[idx].Type = pkg.FieldTypePassword
			case "PIN":
				editedFields[idx].Type = pkg.FieldTypePIN
			default:
				editedFields[idx].Type = pkg.FieldTypeGeneral
			}
		})
		switch field.Type {
		case pkg.FieldTypePassword:
			typeSelect.SetSelected("Password")
		case pkg.FieldTypePIN:
			typeSelect.SetSelected("PIN")
		default:
			typeSelect.SetSelected("General")
		}

		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func(idx int) func() {
			return func() {
				editedFields = append(editedFields[:idx], editedFields[idx+1:]...)
				va.showEditEntryDialog(pkg.Entry{Title: titleEntry.Text, Fields: editedFields}, groupPath)
			}
		}(idx))

		fieldsContainer.Add(container.NewVBox(
			widget.NewLabel("Field Name:"),
			fieldNameEntry,
			widget.NewLabel("Value:"),
			fieldValueEntry,
			container.NewBorder(nil, nil, widget.NewLabel("Type:"), deleteBtn, typeSelect),
			widget.NewSeparator(),
		))
	}

	addFieldBtn := widget.NewButton("Add Field", func() {
		editedFields = append(editedFields, pkg.EntryField{
			Name:  "New Field",
			Value: "",
			Type:  pkg.FieldTypeGeneral,
		})
		va.showEditEntryDialog(pkg.Entry{Title: titleEntry.Text, Fields: editedFields}, groupPath)
	})

	scrollFields := container.NewScroll(fieldsContainer)
	scrollFields.SetMinSize(fyne.NewSize(400, 300))

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Title*:"),
			titleEntry,
			widget.NewSeparator(),
			addFieldBtn,
		),
		nil, nil, nil,
		scrollFields,
	)

	d := dialog.NewCustomConfirm("Edit Entry", "Save", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}

		if titleEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("title cannot be empty"), va.mainWindow)
			return
		}

		updatedEntry := pkg.Entry{
			Title:  titleEntry.Text,
			Fields: editedFields,
		}

		entryPath := pkg.Path{
			GroupIDs: groupPath.GroupIDs,
			EntryID:  entry.ID,
		}

		if err := va.service.UpdateEntry(entryPath, updatedEntry); err != nil {
			dialog.ShowError(fmt.Errorf("error updating entry: %v", err), va.mainWindow)
			return
		}

		if err := va.service.Save(); err != nil {
			dialog.ShowError(fmt.Errorf("error saving: %v", err), va.mainWindow)
			return
		}

		dialog.ShowInformation("Success", "Entry updated successfully!", va.mainWindow)
		va.refreshTree()
	}, va.mainWindow)

	d.Resize(fyne.NewSize(500, 600))
	d.Show()
}

func (va *VaultApp) confirmDeleteGroup() {
	dialog.ShowConfirm("Confirm Delete",
		"Are you sure you want to delete this group and all its contents?",
		func(ok bool) {
			if !ok {
				return
			}

			if err := va.service.DeleteGroup(va.currentPath); err != nil {
				dialog.ShowError(fmt.Errorf("error deleting group: %v", err), va.mainWindow)
				return
			}

			if err := va.service.Save(); err != nil {
				dialog.ShowError(fmt.Errorf("error saving: %v", err), va.mainWindow)
				return
			}

			dialog.ShowInformation("Deleted", "Group deleted successfully", va.mainWindow)
			va.navigateBack()
			va.refreshTree()
		}, va.mainWindow)
}

func (va *VaultApp) confirmDeleteEntry(entry pkg.Entry, groupPath pkg.Path) {
	dialog.ShowConfirm("Confirm Delete",
		fmt.Sprintf("Are you sure you want to delete entry '%s'?", entry.Title),
		func(ok bool) {
			if !ok {
				return
			}

			entryPath := pkg.Path{
				GroupIDs: groupPath.GroupIDs,
				EntryID:  entry.ID,
			}

			if err := va.service.DeleteEntry(entryPath); err != nil {
				dialog.ShowError(fmt.Errorf("error deleting entry: %v", err), va.mainWindow)
				return
			}

			if err := va.service.Save(); err != nil {
				dialog.ShowError(fmt.Errorf("error saving: %v", err), va.mainWindow)
				return
			}

			dialog.ShowInformation("Deleted", "Entry deleted successfully", va.mainWindow)
			va.refreshTree()
			va.showGroupDetails(nil)
		}, va.mainWindow)
}

func (va *VaultApp) showSearchDialog() {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search term...")

	resultsList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {},
	)

	var searchResults []struct {
		entry pkg.Entry
		path  pkg.Path
	}

	searchEntry.OnChanged = func(s string) {
		if s == "" {
			searchResults = nil
			resultsList.Length = func() int { return 0 }
			resultsList.Refresh()
			return
		}

		searchResults = nil
		searchTerm := strings.ToLower(s)

		va.service.TraverseForward(func(info pkg.PathInfo) bool {
			if info.IsEntry && info.Entry != nil {
				if strings.Contains(strings.ToLower(info.Entry.Title), searchTerm) {
					searchResults = append(searchResults, struct {
						entry pkg.Entry
						path  pkg.Path
					}{*info.Entry, info.Path})
					return true
				}
				for _, field := range info.Entry.Fields {
					if strings.Contains(strings.ToLower(field.Name), searchTerm) ||
						strings.Contains(strings.ToLower(field.Value), searchTerm) {
						searchResults = append(searchResults, struct {
							entry pkg.Entry
							path  pkg.Path
						}{*info.Entry, info.Path})
						return true
					}
				}
			}
			return true
		})

		resultsList.Length = func() int { return len(searchResults) }
		resultsList.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(searchResults) {
				obj.(*widget.Label).SetText(searchResults[id].entry.Title)
			}
		}
		resultsList.Refresh()
	}

	resultsList.OnSelected = func(id widget.ListItemID) {
		if id < len(searchResults) {
			result := searchResults[id]
			va.currentPath = pkg.Path{GroupIDs: result.path.GroupIDs}
			va.showEntryDetails(result.entry, result.path)
			va.refreshUI()
		}
	}

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Search Entries"),
			searchEntry,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		resultsList,
	)

	d := dialog.NewCustom("Search", "Close", content, va.mainWindow)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

func (va *VaultApp) performSearch(searchTerm string) {
	// This is called from the main search bar
	// You could implement inline search results here
}

func (va *VaultApp) saveVault() {
	if err := va.service.Save(); err != nil {
		dialog.ShowError(fmt.Errorf("error saving: %v", err), va.mainWindow)
		return
	}
	dialog.ShowInformation("Saved", "Vault saved successfully!", va.mainWindow)
}

func (va *VaultApp) lockVault() {
	dialog.ShowConfirm("Lock Vault",
		"Are you sure you want to lock the vault?",
		func(ok bool) {
			if ok {
				va.service = nil
				va.currentPath = pkg.Path{GroupIDs: []string{}}
				va.showUnlockScreen()
			}
		}, va.mainWindow)
}

func (va *VaultApp) navigateBack() {
	if len(va.currentPath.GroupIDs) > 0 {
		va.currentPath = pkg.GetParentPath(va.currentPath)
		va.refreshUI()
	}
}

func (va *VaultApp) refreshTree() {
	va.treeWidget.Refresh()
	va.updateBreadcrumbs()
	va.updateStatus()
}

func (va *VaultApp) refreshUI() {
	va.refreshTree()
	if len(va.currentPath.GroupIDs) == 0 {
		va.showGroupDetails(nil)
	} else {
		group, err := pkg.FindGroupByPath(va.service.GetWallet(), va.currentPath)
		if err == nil {
			va.showGroupDetails(group)
		}
	}
}

func (va *VaultApp) updateBreadcrumbs() {
	va.breadcrumbs.SetText(va.getBreadcrumbText())
}

func (va *VaultApp) updateStatus() {
	va.statusLabel.SetText(va.getStatusText())
}

func (va *VaultApp) getBreadcrumbText() string {
	if len(va.currentPath.GroupIDs) == 0 {
		return "📁 ROOT"
	}

	breadcrumbs := []string{"ROOT"}
	currentPath := pkg.Path{GroupIDs: []string{}}

	for _, groupID := range va.currentPath.GroupIDs {
		currentPath.GroupIDs = append(currentPath.GroupIDs, groupID)
		group, err := pkg.FindGroupByPath(va.service.GetWallet(), currentPath)
		if err == nil {
			breadcrumbs = append(breadcrumbs, group.Name)
		}
	}

	return "📁 " + strings.Join(breadcrumbs, " > ")
}

func (va *VaultApp) getStatusText() string {
	wallet := va.service.GetWallet()
	if wallet == nil {
		return "No wallet loaded"
	}

	totalGroups := 0
	totalEntries := 0

	va.service.TraverseForward(func(info pkg.PathInfo) bool {
		if info.IsEntry {
			totalEntries++
		} else {
			totalGroups++
		}
		return true
	})

	return fmt.Sprintf("Vault unlocked | Groups: %d | Entries: %d", totalGroups, totalEntries)
}

func generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

func main() {
	app := NewVaultApp()
	app.Run()
}
