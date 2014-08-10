VERSION := $(shell cat main.go | grep "const VERSION" | awk -F "\"" '{print $$2}')

rpm: clean
	mkdir -p build/{BUILD,RPMS,SOURCES,SPECS,SRPMS,tmp}
	mkdir build/BUILD/emftoimg-$(VERSION)
	cp -r *.go emf Makefile build/BUILD/emftoimg-$(VERSION)
	cp -r contrib/emftoimg.spec build/SPECS/
	rpmbuild --define "_topdir `pwd`/build" --define "_tmppath `pwd`/build/tmp" \
		--define "version $(VERSION)" -bb build/SPECS/emftoimg.spec

clean:
	rm -rf build/
