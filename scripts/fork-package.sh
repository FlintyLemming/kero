#!/usr/bin/env bash
# fork-package.sh builds this fork's Kero for Apple Silicon, ad hoc signed
# (no Apple developer account: Apple Silicon runs nothing less than that),
# into build/fork/: Kero-<version>.dmg and .zip, and the app itself.
#
# The version is <upstream version>-patch-<yymmdd>, and the build number
# <yymmdd>.<HHMM>, which only grows, as Sparkle needs. Both are passed to
# xcodebuild, not committed. DERIVED_DATA reuses a build directory.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
out=build/fork
dd="${DERIVED_DATA:-build/DerivedData}"
base="$(sed -n 's/.*MARKETING_VERSION = \([0-9.]*\);.*/\1/p' kero.xcodeproj/project.pbxproj | head -1)"
[ -n "$base" ] || { echo "fork-package: no MARKETING_VERSION in the project" >&2; exit 1; }
version="$base-patch-$(date +%y%m%d)"
build="$(date +%y%m%d.%H%M)"

git submodule update --init --recursive
echo "fork-package: building Kero $version ($build)"
# -scmProvider system fetches the packages with the system git: Xcode's
# own asks the keychain for github.com's password.
# No hardened runtime: it only matters for notarization, and an ad hoc
# app with it can be refused its own ad hoc frameworks.
xcodebuild -project kero.xcodeproj -scheme kero -configuration Release \
	-destination 'platform=macOS,arch=arm64' -derivedDataPath "$dd" -scmProvider system \
	CODE_SIGN_IDENTITY=- CODE_SIGN_STYLE=Manual DEVELOPMENT_TEAM= PROVISIONING_PROFILE_SPECIFIER= \
	ENABLE_HARDENED_RUNTIME=NO MARKETING_VERSION="$version" CURRENT_PROJECT_VERSION="$build" \
	build | grep -E '^\*\* BUILD|error: ' || true

app="$dd/Build/Products/Release/Kero.app"
[ -d "$app" ] || { echo "fork-package: no $app; see the build output above" >&2; exit 1; }
[ "$(plutil -extract CFBundleVersion raw "$app/Contents/Info.plist")" = "$build" ] ||
	{ echo "fork-package: the build failed; $app is an older one" >&2; exit 1; }
codesign --verify --deep --strict "$app"

rm -rf "$out"
mkdir -p "$out/stage"
ditto "$app" "$out/Kero.app"
ditto "$app" "$out/stage/Kero.app"
ln -s /Applications "$out/stage/Applications"
hdiutil create -quiet -volname "Kero $version" -srcfolder "$out/stage" -ov -format UDZO "$out/Kero-$version.dmg"
rm -rf "$out/stage"
ditto -c -k --keepParent "$app" "$out/Kero-$version.zip"
echo "fork-package: built into $out:"
ls -lh "$out" | sed 1d
