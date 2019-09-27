package updater

// Default file names
const (
	CLIENT_WYC      = "client.wyc"
	IUCLIENT_IUC    = "iuclient.iuc"    // inside client.wyc
	UPDTDETAILS_UDT = "updtdetails.udt" // inside .wyu archive

)

// File headers
const (
	IUC_HEADER         = "IUCDFV2"
	WYS_HEADER         = "IUSDFV2"
	UPDTDETAILS_HEADER = "IUUDFV2"
)

// Exit codes
const (
	EXIT_SUCCESS          = 0
	EXIT_NO_UPDATE        = 0
	EXIT_ERROR            = 1
	EXIT_UPDATE_AVALIABLE = 2
)
