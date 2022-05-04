# Download and unzip wintun libraries

mkdir -p artifacts
wget -O artifacts/wintun.zip https://www.wintun.net/builds/wintun-0.14.1.zip

rm -fr artifacts/Windows
unzip -d artifacts/Windows artifacts/wintun.zip

mkdir artifacts/Windows/{i386,x86_64,arm_64}
mv artifacts/Windows/wintun/bin/amd64/wintun.dll artifacts/Windows/x86_64/wintun.dll
mv artifacts/Windows/wintun/bin/arm64/wintun.dll artifacts/Windows/arm_64/wintun.dll
mv artifacts/Windows/wintun/bin/x86/wintun.dll artifacts/Windows/i386/wintun.dll
rm -fr artifacts/Windows/wintun artifacts/wintun.zip