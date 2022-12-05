(define-module (nordic-channel packages misc)
  #:use-module (guix build-system go)
  #:use-module (guix git-download)
  #:use-module ((guix licenses) #:prefix license:)
  #:use-module (guix packages))
;(define-public ppm
  (package
   (name "pmanager-go")
   (version "0.1")
   (source (origin
	    (method git-fetch)
	    (uri (git-reference
		  (url "https://github.com/SMproductive/ppm")
		  (commit "473765e44b105da132bae30b708faad81fce0c24")))
	    (file-name (git-file-name name version))
	    (sha256
	     (base32
	      "17cm9s9pm00hgg3vafrq9gc60cvk52cxgs6paqfwdz76ljp1b6k8"))))
   (build-system go-build-system)
   (native-inputs
    (list))
   (arguments
    `(#:import-path "github.com/SMproductive/ppm"))
   (home-page "https://github.com/SMproductive/ppm")
   (synopsis "Piping Password Manger")
   (description "Piping Password Manager manages your passwords and to prevent decryption over and over again you can use pipes to communicate with it.")
   (license license:gpl3+));)
