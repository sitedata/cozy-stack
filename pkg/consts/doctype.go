package consts

// Instances doc type for User's instance document
const Instances = "instances"

// Configs doc type assets documents configuration
const Configs = "configs"

const (
	// Apps doc type for client-side application manifests
	Apps = "io.cozy.apps"
	// AppsSuggestion doc type for suggesting apps to the user
	AppsSuggestion = "io.cozy.apps.suggestions"
	// Konnectors doc type for konnector application manifests
	Konnectors = "io.cozy.konnectors"
	// KonnectorsMaintenance doc type for maintenance of konnectors.
	KonnectorsMaintenance = "io.cozy.konnectors.maintenance"
	// Archives doc type for zip archives with files and directories
	Archives = "io.cozy.files.archives"
	// Exports doc type for global exports archives
	Exports = "io.cozy.exports"
	// ExportsRequests doc type for a request to move to another Cozy
	ExportsRequests = "io.cozy.exports.requests"
	// Imports doc type for global exports archives
	Imports = "io.cozy.imports"
	// Doctypes doc type for doctype list
	Doctypes = "io.cozy.doctypes"
	// Files doc type for type for files and directories
	Files = "io.cozy.files"
	// FilesMetadata doc type for metadata of files
	FilesMetadata = "io.cozy.files.metadata"
	// FilesVersions doc type for versioning file contents
	FilesVersions = "io.cozy.files.versions"
	// FilesShortcuts doc type for high-level information about .url files
	FilesShortcuts = "io.cozy.files.shortcuts"
	// Thumbnails is a synthetic doctype for thumbnails, used for realtime
	// events
	Thumbnails = "io.cozy.files.thumbnails"
	// CertifiedCarbonCopy is a synthetic doctype, used for given permission to
	// add the carbonCopy metadata on files
	CertifiedCarbonCopy = "io.cozy.certified.carbon_copy"
	// CertifiedElectronicSafe is a synthetic doctype, used for given
	// permission to add the electronicSafe metadata on files
	CertifiedElectronicSafe = "io.cozy.certified.electronic_safe"
	// PhotosAlbums doc type for photos albums
	PhotosAlbums = "io.cozy.photos.albums"
	// Intents doc type for intents persisted in couchdb
	Intents = "io.cozy.intents"
	// Jobs doc type for queued jobs
	Jobs = "io.cozy.jobs"
	// JobEvents doc type for real time events sent by jobs
	JobEvents = "io.cozy.jobs.events"
	// Support doc type for sending mail to the support
	Support = "io.cozy.support"
	// Notifications doc type for notifications
	Notifications = "io.cozy.notifications"
	// OAuthAccessCodes doc type for OAuth2 access codes
	OAuthAccessCodes = "io.cozy.oauth.access_codes"
	// OAuthClients doc type for OAuth2 clients
	OAuthClients = "io.cozy.oauth.clients"
	// Permissions doc type for permissions identifying a connection
	Permissions = "io.cozy.permissions"
	// Contacts doc type for sharing
	Contacts = "io.cozy.contacts"
	// RemoteRequests doc type for logging requests to remote websites
	RemoteRequests = "io.cozy.remote.requests"
	// RemoteSecrets doc type for secrets used by remote doctypes
	RemoteSecrets = "io.cozy.remote.secrets"
	// Sessions doc type for sessions identifying a connection
	Sessions = "io.cozy.sessions"
	// SessionsLogins doc type for sessions identifying a connection
	SessionsLogins = "io.cozy.sessions.logins"
	// Settings doc type for settings to customize an instance
	Settings = "io.cozy.settings"
	// Shared doc type for keepking track of documents in sharings
	Shared = "io.cozy.shared"
	// Sharings doc type for document and file sharing
	Sharings = "io.cozy.sharings"
	// SharingsMembers doc type for members of a sharing
	SharingsMembers = "io.cozy.sharings.members"
	// SharingsAnswer doc type for credentials exchange for sharings
	SharingsAnswer = "io.cozy.sharings.answer"
	// SharingsMoved doc type for when a Cozy is moved to a new address
	SharingsMoved = "io.cozy.sharings.moved"
	// SharingsInitialSync doc type for real-time events for initial sync of a
	// sharing
	SharingsInitialSync = "io.cozy.sharings.initial_sync"
	// Triggers doc type for triggers, jobs launchers
	Triggers = "io.cozy.triggers"
	// TriggersState doc type for triggers current state, jobs launchers
	TriggersState = "io.cozy.triggers.state"
	// Accounts doc type for accounts
	Accounts = "io.cozy.accounts"
	// AccountTypes doc type for account types
	AccountTypes = "io.cozy.account_types"
	// BitwardenProfiles doc type for Bitwarden profile
	BitwardenProfiles = "com.bitwarden.profiles"
	// BitwardenCiphers doc type for Bitwarden ciphers
	BitwardenCiphers = "com.bitwarden.ciphers"
	// BitwardenFolders doc type for Bitwarden folders
	BitwardenFolders = "com.bitwarden.folders"
	// BitwardenOrganizations doc type for Bitwarden organizations
	BitwardenOrganizations = "com.bitwarden.organizations"
	// BitwardenContacts doc type for Bitwarden users that can be added to
	// an organization
	BitwardenContacts = "com.bitwarden.contacts"
	// NotesDocuments doc type is used for manipulating the documents that
	// represents a note before they are persisted to a file.
	NotesDocuments = "io.cozy.notes.documents"
	// NotesSteps doc type is used for patching a note.
	NotesSteps = "io.cozy.notes.steps"
	// NotesTelepointers doc type is used for the position of the cursor in a
	// note.
	NotesTelepointers = "io.cozy.notes.telepointers"
	// NotesEvents doc type is used for realtime events related to a note, like
	// a change of title.
	NotesEvents = "io.cozy.notes.events"
	// NotesURL doc type is used to return the URL where a note can be edited.
	NotesURL = "io.cozy.notes.url"
	// NotesImages doc type used for images used by a note
	NotesImages = "io.cozy.notes.images"
	// OfficeURL doc type is used to return the URL where an office document can be edited.
	OfficeURL = "io.cozy.office.url"
	// AuthConfirmations doc type used for realtime events when confirming
	// authentication.
	AuthConfirmations = "io.cozy.auth.confirmations"
)
