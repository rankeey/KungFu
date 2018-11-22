CMAKE_MINIMUM_REQUIRED(VERSION 3.5)
PROJECT(kungfu)

SET(CMAKE_CXX_STANDARD 17)

INCLUDE(cmake/tf-op.cmake)

ADD_TF_OP_LIB(kungfu_tensorflow_ops
              src/negotiator.cpp
              src/communicator.cpp
              src/message.cpp)

ENABLE_TESTING()
# TODO: add tests