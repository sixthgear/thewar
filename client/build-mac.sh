SRC_PATH=`pwd`
APP_PATH=../build/THEWARCLIENT.app
HASH=`git rev-parse --short HEAD`

# do actual build
go build

# remove old App
rm -rf $APP_PATH

# create barebones Application structure
mkdir -p $APP_PATH/Contents/MacOS
mkdir -p $APP_PATH/Contents/Resources
cp $SRC_PATH/mac/Info.plist $APP_PATH/Contents/
cp $SRC_PATH/mac/icon.icns $APP_PATH/Contents/Resources

# copy binary
cp $SRC_PATH/client $APP_PATH/Contents/MacOS/THEWARCLIENT

# copy data files
cp -R $SRC_PATH/data $APP_PATH/Contents/Resources

# copy required shared libs
cp /usr/local/Cellar/glew/1.10.0/lib/libGLEW.1.10.0.dylib $APP_PATH/Contents/MacOS
# cp /usr/local/Cellar/glfw3/3.0.3/lib/libglfw3.a $APP_PATH/Contents/MacOS

# relink binary to point to shared libs in local path
install_name_tool -change /usr/local/lib/libGLEW.1.10.0.dylib  @executable_path/libGLEW.1.10.0.dylib $APP_PATH/Contents/MacOS/THEWARCLIENT
# install_name_tool -change /usr/local/opt/glfw/lib/libglfw.dylib  @executable_path/libglfw.dylib $APP_PATH/Contents/MacOS/THEWARCLIENT
tar -C ../build -czf ~/Dropbox/Public/THEWAR/THEWARCLIENT-${HASH}.tar.gz THEWARCLIENT.app
