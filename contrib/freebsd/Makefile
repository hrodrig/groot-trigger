PORTNAME=	groot-trigger
PORTVERSION=	0.1.2
DISTVERSIONPREFIX=	v
CATEGORIES=	sysutils
# Local test before release: copy the tarball into DISTDIR, or run e.g.
#   make fetch MASTER_SITES=file:///absolute/path/to/dir/
# See README.md "Test with a local distfile".
MASTER_SITES=	https://github.com/hrodrig/groot-trigger/releases/download/v${PORTVERSION}/
DISTFILES=	groot-trigger_v${PORTVERSION}_freebsd_${ARCH:S/aarch64/arm64/}.tar.gz
EXTRACT_SUFX=	.tar.gz

MAINTAINER=	hrodrig@usb.ve
COMMENT=	On-demand HTTP API that creates Kubernetes Jobs running groot collect
WWW=		https://github.com/hrodrig/groot-trigger
LICENSE=	MIT

NO_BUILD=	yes
ONLY_FOR_ARCHS=	amd64 aarch64

do-install:
	${INSTALL_PROGRAM} ${WRKDIR}/groot-trigger ${STAGEDIR}${PREFIX}/bin/groot-trigger
	${MKDIR} ${STAGEDIR}${PREFIX}/share/man/man1
	${INSTALL_MAN} ${WRKDIR}/share/man/man1/groot-trigger.1 ${STAGEDIR}${PREFIX}/share/man/man1/
	${MKDIR} ${STAGEDIR}${PREFIX}/share/doc/groot-trigger
	${INSTALL_DATA} ${WRKDIR}/share/doc/groot-trigger/LICENSE ${STAGEDIR}${PREFIX}/share/doc/groot-trigger/
	${INSTALL_DATA} ${WRKDIR}/share/doc/groot-trigger/README.md ${STAGEDIR}${PREFIX}/share/doc/groot-trigger/

.include <bsd.port.mk>
