# Truco AI Engine — C++ shared library built via CGO
# This Makefile builds libtruco_ai.so for the current platform

CXX ?= g++
CXXFLAGS = -std=c++17 -O2 -fPIC -Wall -Wextra -Iinclude
SRCS = src/Card.cpp src/GameState.cpp src/Strategy.cpp src/AIEngine.cpp src/truco_ai_c.cpp
OBJS = $(SRCS:.cpp=.o)
TARGET = libtruco_ai.so

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CXX) -shared $(CXXFLAGS) -o $@ $(OBJS)

%.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $@

clean:
	rm -f $(OBJS) $(TARGET)
