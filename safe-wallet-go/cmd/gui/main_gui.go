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
	va.mainWindow = va.app.NewWindow("Safe Wallet Go! - Password Vault")
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

		createVault := func() {
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
		}

		createBtn := widget.NewButton("Create Vault", createVault)
		createBtn.Importance = widget.HighImportance

		// Allow Enter key to submit
		passwordEntry.OnSubmitted = func(s string) {
			confirmEntry.FocusGained()
		}
		confirmEntry.OnSubmitted = func(s string) {
			createVault()
		}

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

		// Set focus on password entry after a short delay
		va.mainWindow.Canvas().Focus(passwordEntry)
	} else {
		// Unlock existing wallet
		subtitle := widget.NewLabel("Enter your master password to unlock")
		subtitle.Alignment = fyne.TextAlignCenter

		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.SetPlaceHolder("Master Password")

		unlockVault := func() {
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
		}

		unlockBtn := widget.NewButton("Unlock Vault", unlockVault)
		unlockBtn.Importance = widget.HighImportance

		// Allow Enter key to submit
		passwordEntry.OnSubmitted = func(s string) {
			unlockVault()
		}

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

		// Set focus on password entry
		va.mainWindow.Canvas().Focus(passwordEntry)
	}

	va.mainWindow.SetContent(content)
}

func (va *VaultApp) showMainInterface() {
	// Create toolbar
	toolbar := va.createToolbar()

	// Create breadcrumbs
	va.breadcrumbs = widget.NewLabel(va.getBreadcrumbText())
	va.breadcrumbs.TextStyle = fyne.TextStyle{Bold: true}

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
		va.breadcrumbs,
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
			va.treeWidget.UnselectAll()
			va.treeWidget.CloseAllBranches()
			va.showGroupDetails(nil)
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
	// Update current path and breadcrumbs when viewing an entry
	va.currentPath = groupPath
	va.updateBreadcrumbs()

	title := widget.NewLabelWithStyle(entry.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapWord

	fieldsContainer := container.NewVBox()

	for _, field := range entry.Fields {
		fieldLabel := widget.NewLabel(field.Name + ":")
		fieldLabel.TextStyle = fyne.TextStyle{Bold: true}

		var valueWidget fyne.CanvasObject

		if field.Type == pkg.FieldTypePassword || field.Type == pkg.FieldTypePIN {
			// Use a label instead of entry for better scrolling
			valueLabel := widget.NewLabel(field.Value)
			if field.Type == pkg.FieldTypePassword || field.Type == pkg.FieldTypePIN {
				// Show as dots initially
				valueLabel.SetText(strings.Repeat("•", len(field.Value)))
			}

			showBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), func(lbl *widget.Label, val string, isHidden *bool) func() {
				hidden := true
				isHidden = &hidden
				return func() {
					if *isHidden {
						lbl.SetText(val)
						*isHidden = false
					} else {
						lbl.SetText(strings.Repeat("•", len(val)))
						*isHidden = true
					}
				}
			}(valueLabel, field.Value, nil))

			copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func(val string, name string) func() {
				return func() {
					va.mainWindow.Clipboard().SetContent(val)
					dialog.ShowInformation("Copied", name+" copied to clipboard", va.mainWindow)
				}
			}(field.Value, field.Name))

			valueWidget = container.NewBorder(nil, nil, nil,
				container.NewHBox(showBtn, copyBtn), valueLabel)
		} else {
			// For general fields, use label for better scrolling
			valueLabel := widget.NewLabel(field.Value)
			valueLabel.Wrapping = fyne.TextWrapWord

			copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func(val string, name string) func() {
				return func() {
					va.mainWindow.Clipboard().SetContent(val)
					dialog.ShowInformation("Copied", name+" copied to clipboard", va.mainWindow)
				}
			}(field.Value, field.Name))

			valueWidget = container.NewBorder(nil, nil, nil, copyBtn, valueLabel)
		}

		fieldsContainer.Add(fieldLabel)
		fieldsContainer.Add(valueWidget)
		fieldsContainer.Add(widget.NewSeparator())
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
		widget.NewSeparator(),
		buttons,
	)

	scroll := container.NewScroll(details)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	va.detailsPanel.Objects = []fyne.CanvasObject{scroll}
	va.detailsPanel.Refresh()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	var fieldEntries []*widget.Entry

	updateFieldsUI := func(templateName string) {
		fieldsContainer.Objects = nil
		fields = []pkg.EntryField{}
		fieldEntries = []*widget.Entry{}

		if templateName == "Custom" {
			// Show add field button for custom
			addFieldBtn := widget.NewButton("Add Field", func() {
				va.showAddCustomFieldDialog(&fields, fieldsContainer, &fieldEntries)
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

						if templateField.Type == pkg.FieldTypePassword {
							entry = widget.NewPasswordEntry()
						} else if templateField.Type == pkg.FieldTypePIN {
							entry = widget.NewEntry()
							entry.SetPlaceHolder("Numbers only")
							idx := len(fields) - 1
							entry.OnChanged = func(idx int) func(string) {
								return func(s string) {
									filtered := ""
									for _, r := range s {
										if r >= '0' && r <= '9' {
											filtered += string(r)
										}
									}
									if filtered != s {
										entry.SetText(filtered)
									}
									fields[idx].Value = filtered
								}
							}(idx)
						} else {
							entry = widget.NewEntry()
							idx := len(fields) - 1
							entry.OnChanged = func(idx int) func(string) {
								return func(s string) {
									fields[idx].Value = s
								}
							}(idx)
						}

						if templateField.Type != pkg.FieldTypePIN {
							idx := len(fields) - 1
							entry.OnChanged = func(idx int) func(string) {
								return func(s string) {
									fields[idx].Value = s
								}
							}(idx)
						}

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

func (va *VaultApp) showAddCustomFieldDialog(fields *[]pkg.EntryField, fieldsContainer *fyne.Container, fieldEntries *[]*widget.Entry) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Field Name")

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Field Value")

	typeSelect := widget.NewSelect([]string{"General", "Password", "PIN"}, nil)
	typeSelect.SetSelected("General")

	typeSelect.OnChanged = func(selected string) {
		if selected == "PIN" {
			valueEntry.SetPlaceHolder("Numbers only")
			valueEntry.OnChanged = func(s string) {
				filtered := ""
				for _, r := range s {
					if r >= '0' && r <= '9' {
						filtered += string(r)
					}
				}
				if filtered != s {
					valueEntry.SetText(filtered)
				}
			}
		} else {
			valueEntry.SetPlaceHolder("Field Value")
			valueEntry.OnChanged = nil
		}
	}

	d := dialog.NewForm("Add Field", "Add", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name*", nameEntry),
		widget.NewFormItem("Value*", valueEntry),
		widget.NewFormItem("Type", typeSelect),
	}, func(ok bool) {
		if !ok {
			return
		}

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
			if !pkg.IsNumeric(valueEntry.Text) {
				dialog.ShowError(fmt.Errorf("PIN must contain only numbers"), va.mainWindow)
				return
			}
		}

		newField := pkg.EntryField{
			Name:  nameEntry.Text,
			Value: valueEntry.Text,
			Type:  fieldType,
		}
		*fields = append(*fields, newField)

		// Add field display to UI
		fieldBox := container.NewHBox()
		fieldLabel := widget.NewLabel(nameEntry.Text + ": ")
		fieldLabel.TextStyle = fyne.TextStyle{Bold: true}

		var valueLabel *widget.Label
		if fieldType == pkg.FieldTypePassword || fieldType == pkg.FieldTypePIN {
			valueLabel = widget.NewLabel(strings.Repeat("•", len(valueEntry.Text)))
		} else {
			valueLabel = widget.NewLabel(valueEntry.Text)
		}

		// Add delete button for the field
		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func(idx int) func() {
			return func() {
				// Remove from fields slice
				*fields = append((*fields)[:idx], (*fields)[idx+1:]...)
				// Rebuild UI
				va.rebuildCustomFieldsUI(fields, fieldsContainer)
			}
		}(len(*fields)-1))

		fieldBox.Add(fieldLabel)
		fieldBox.Add(valueLabel)
		fieldBox.Add(deleteBtn)

		// Insert before the "Add Field" button
		if len(fieldsContainer.Objects) > 0 {
			fieldsContainer.Objects = append(fieldsContainer.Objects[:len(fieldsContainer.Objects)-1], fieldBox)
			fieldsContainer.Objects = append(fieldsContainer.Objects, fieldsContainer.Objects[len(fieldsContainer.Objects)-1])
			fieldsContainer.Objects[len(fieldsContainer.Objects)-2] = fieldBox
		}
		fieldsContainer.Refresh()
	}, va.mainWindow)

	d.Resize(fyne.NewSize(400, 250))
	d.Show()
}

func (va *VaultApp) rebuildCustomFieldsUI(fields *[]pkg.EntryField, fieldsContainer *fyne.Container) {
	// Keep only the "Add Field" button
	addFieldBtn := fieldsContainer.Objects[0]
	fieldsContainer.Objects = []fyne.CanvasObject{addFieldBtn}

	// Rebuild all field displays
	for i, field := range *fields {
		fieldBox := container.NewHBox()
		fieldLabel := widget.NewLabel(field.Name + ": ")
		fieldLabel.TextStyle = fyne.TextStyle{Bold: true}

		var valueLabel *widget.Label
		if field.Type == pkg.FieldTypePassword || field.Type == pkg.FieldTypePIN {
			valueLabel = widget.NewLabel(strings.Repeat("•", len(field.Value)))
		} else {
			valueLabel = widget.NewLabel(field.Value)
		}

		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func(idx int) func() {
			return func() {
				*fields = append((*fields)[:idx], (*fields)[idx+1:]...)
				va.rebuildCustomFieldsUI(fields, fieldsContainer)
			}
		}(i))

		fieldBox.Add(fieldLabel)
		fieldBox.Add(valueLabel)
		fieldBox.Add(deleteBtn)

		fieldsContainer.Objects = append(fieldsContainer.Objects[:len(fieldsContainer.Objects)-1], fieldBox, addFieldBtn)
	}

	fieldsContainer.Refresh()
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

	// Store the original entry ID
	originalEntryID := entry.ID
	var rebuildFieldsUI func()
	rebuildFieldsUI = func() {
		fieldsContainer.Objects = nil

		for i, field := range editedFields {
			idx := i

			fieldBox := container.NewVBox()

			fieldNameLabel := widget.NewLabel("Field Name:")
			fieldNameEntry := widget.NewEntry()
			fieldNameEntry.SetText(field.Name)
			fieldNameEntry.OnChanged = func(s string) {
				editedFields[idx].Name = s
			}

			valueLabel := widget.NewLabel("Value:")
			var fieldValueEntry *widget.Entry
			if field.Type == pkg.FieldTypePassword {
				fieldValueEntry = widget.NewPasswordEntry()
			} else if field.Type == pkg.FieldTypePIN {
				fieldValueEntry = widget.NewEntry()
				fieldValueEntry.SetPlaceHolder("Numbers only")
				fieldValueEntry.OnChanged = func(s string) {
					filtered := ""
					for _, r := range s {
						if r >= '0' && r <= '9' {
							filtered += string(r)
						}
					}
					if filtered != s {
						fieldValueEntry.SetText(filtered)
					}
					editedFields[idx].Value = filtered
				}
			} else {
				fieldValueEntry = widget.NewEntry()
				fieldValueEntry.OnChanged = func(s string) {
					editedFields[idx].Value = s
				}
			}
			fieldValueEntry.SetText(field.Value)

			typeSelect := widget.NewSelect([]string{"General", "Password", "PIN"}, func(s string) {
				switch s {
				case "Password":
					editedFields[idx].Type = pkg.FieldTypePassword
				case "PIN":
					editedFields[idx].Type = pkg.FieldTypePIN
					if !pkg.IsNumeric(editedFields[idx].Value) {
						editedFields[idx].Value = ""
						fieldValueEntry.SetText("")
					}
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

			deleteBtn := widget.NewButtonWithIcon("Delete Field", theme.DeleteIcon(), func(idx int) func() {
				return func() {
					editedFields = append(editedFields[:idx], editedFields[idx+1:]...)
					rebuildFieldsUI()
				}
			}(idx))
			deleteBtn.Importance = widget.DangerImportance

			fieldBox.Add(fieldNameLabel)
			fieldBox.Add(fieldNameEntry)
			fieldBox.Add(valueLabel)
			fieldBox.Add(fieldValueEntry)
			fieldBox.Add(container.NewBorder(nil, nil, widget.NewLabel("Type:"), nil, typeSelect))
			fieldBox.Add(deleteBtn)
			fieldBox.Add(widget.NewSeparator())

			fieldsContainer.Add(fieldBox)
		}
		fieldsContainer.Refresh()
	}

	rebuildFieldsUI()

	addFieldBtn := widget.NewButton("Add Field", func() {
		editedFields = append(editedFields, pkg.EntryField{
			Name:  "New Field",
			Value: "",
			Type:  pkg.FieldTypeGeneral,
		})
		rebuildFieldsUI()
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

		for _, field := range editedFields {
			if field.Type == pkg.FieldTypePIN && !pkg.IsNumeric(field.Value) {
				dialog.ShowError(fmt.Errorf("PIN field '%s' must contain only numbers", field.Name), va.mainWindow)
				return
			}
		}

		updatedEntry := pkg.Entry{
			ID:     originalEntryID,
			Title:  titleEntry.Text,
			Fields: editedFields,
		}

		entryPath := pkg.Path{
			GroupIDs: groupPath.GroupIDs,
			EntryID:  originalEntryID,
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
		va.showEntryDetails(updatedEntry, groupPath)
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
			return container.NewVBox(
				widget.NewLabel("Template Title"),
				widget.NewLabel("Path"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {},
	)

	var searchResults []struct {
		entry pkg.Entry
		path  pkg.Path
	}

	var currentDialog dialog.Dialog

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
				c := obj.(*fyne.Container)
				titleLabel := c.Objects[0].(*widget.Label)
				pathLabel := c.Objects[1].(*widget.Label)

				result := searchResults[id]
				titleLabel.SetText(result.entry.Title)
				titleLabel.TextStyle = fyne.TextStyle{Bold: true}

				// Build path string
				pathParts := []string{"ROOT"}
				currentPath := pkg.Path{GroupIDs: []string{}}
				for _, groupID := range result.path.GroupIDs {
					currentPath.GroupIDs = append(currentPath.GroupIDs, groupID)
					group, err := pkg.FindGroupByPath(va.service.GetWallet(), currentPath)
					if err == nil {
						pathParts = append(pathParts, group.Name)
					}
				}
				pathLabel.SetText("📁 " + strings.Join(pathParts, " > "))
			}
		}
		resultsList.Refresh()
	}

	resultsList.OnSelected = func(id widget.ListItemID) {
		if id < len(searchResults) {
			result := searchResults[id]

			// Close the search dialog
			if currentDialog != nil {
				currentDialog.Hide()
			}

			// Navigate to the entry's parent group
			va.currentPath = pkg.Path{GroupIDs: result.path.GroupIDs}

			// Expand the tree to show the entry
			va.expandTreeToPath(result.path.GroupIDs, result.entry.ID)

			// Show entry details
			va.showEntryDetails(result.entry, result.path)

			// Update UI
			va.updateBreadcrumbs()
			va.refreshTree()
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

	currentDialog = dialog.NewCustom("Search", "Close", content, va.mainWindow)
	currentDialog.Resize(fyne.NewSize(600, 500))
	currentDialog.Show()
}

func (va *VaultApp) expandTreeToPath(groupIDs []string, entryID string) {
	// Build the tree path and open each node
	treePath := ""
	for i, groupID := range groupIDs {
		if i == 0 {
			treePath = groupID
		} else {
			treePath = treePath + "|" + groupID
		}
		va.treeWidget.OpenBranch(treePath)
	}

	// Select the entry in the tree
	if len(groupIDs) > 0 {
		entryPath := strings.Join(groupIDs, "|") + "|E:" + entryID
		va.treeWidget.Select(entryPath)
	}
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

		// Update tree selection to highlight current location
		if len(va.currentPath.GroupIDs) == 0 {
			// At root - unselect all
			va.treeWidget.UnselectAll()
			va.showGroupDetails(nil)
		} else {
			// Select the current group in the tree
			treeID := strings.Join(va.currentPath.GroupIDs, "|")
			va.treeWidget.Select(treeID)

			// Show group details
			group, err := pkg.FindGroupByPath(va.service.GetWallet(), va.currentPath)
			if err == nil {
				va.showGroupDetails(group)
			}
		}

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
