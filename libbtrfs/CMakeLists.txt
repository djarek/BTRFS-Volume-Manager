cmake_minimum_required(VERSION 3.2)
set(CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/CMake" ${CMAKE_MODULE_PATH})
project(btrfs)

set(CMAKE_CC_FLAGS "-std=c99 -Wall -Werror -fno-strict-aliasing")

MESSAGE(STATUS ${CMAKE_CC_FLAGS_RELEASE})
include(btrfs-iface/CMakeLists.txt)
add_library(btrfs SHARED libbtrfs.c ${iface_dir})

target_link_libraries(btrfs blkid)
target_compile_definitions(btrfs PRIVATE BTRFS_FLAT_INCLUDES=1 FORTIFY_SOURCE=2)

project(libbtrfs-test)
add_executable(libbtrfs-test libbtrfs_test.c)
target_link_libraries(libbtrfs-test btrfs)