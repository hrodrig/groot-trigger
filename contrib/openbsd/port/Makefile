# OpenBSD port for groot-trigger
# Copy to /usr/ports/sysutils/groot-trigger/ and run: make makesum
# Submit: send diff to ports@openbsd.org

COMMENT =	on-demand HTTP API creating Kubernetes Jobs for groot collect
DISTNAME =	groot-trigger_v0.1.2_openbsd_${MACHINE_ARCH:S/aarch64/arm64/}
PKGNAME =	groot-trigger-0.1.2
CATEGORIES =	sysutils
HOMEPAGE =	https://github.com/hrodrig/groot-trigger
MAINTAINER =	hrodrig@usb.ve

# MIT
PERMIT_PACKAGE =	Yes
PERMIT_DISTFILES =	Yes

# Local test: copy tarball to DISTDIR or
#   make fetch MASTER_SITES=file:///absolute/path/to/dir/
MASTER_SITES =	https://github.com/hrodrig/groot-trigger/releases/download/v0.1.2/
DISTFILES =	groot-trigger_v0.1.2_openbsd_${MACHINE_ARCH:S/aarch64/arm64/}.tar.gz
EXTRACT_SUFX =	.tar.gz

ONLY_FOR_ARCHS =	amd64 arm64 aarch64
NO_BUILD =	Yes
NO_TEST =	Yes
PKG_ARCH =	*

do-install:
	${INSTALL_PROGRAM} ${WRKDIR}/groot-trigger ${PREFIX}/bin/groot-trigger
	${INSTALL_MAN} ${WRKDIR}/share/man/man1/groot-trigger.1 ${PREFIX}/man/man1/groot-trigger.1
	${INSTALL_DATA_DIR} ${PREFIX}/share/doc/groot-trigger
	${INSTALL_DATA} ${WRKDIR}/share/doc/groot-trigger/LICENSE ${PREFIX}/share/doc/groot-trigger/
	${INSTALL_DATA} ${WRKDIR}/share/doc/groot-trigger/README.md ${PREFIX}/share/doc/groot-trigger/

.include <bsd.port.mk>
