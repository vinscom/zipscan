// Copyright (C) 2015 Vinay Kumar
//
// zipscan is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// zipscan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

package main

import (
	"sync"
)

type fileInfo struct {
	processed         bool
	path              string
	name              string
	dir               bool
	foundContentMatch bool
	foundPathMatch    bool
}

type processorInfo struct {
	fn    processor
	count int
}

type processor func(fileInfoChannel, fileInfoChannel)
type processorFn func(fileInfo, fileInfoChannel)

func NewProcessor(fn processorFn) processor {
	return func(in fileInfoChannel, out fileInfoChannel) {

		if in == nil {
			fn(fileInfo{}, out)
			return
		}

		for {
			select {
			case i, ok := <-in:
				if ok {
					fn(i, out)
				} else {
					return
				}
			}
		}
	}
}

func SetupSystem(p *[]processorInfo) fileInfoChannel {
	var lastProcessorOutChannel fileInfoChannel = nil
	for _, e := range *p {
		newOutChannel := make(fileInfoChannel,100)
		if e.count > 1 {
			go NewProcessorRunnerWithWait(lastProcessorOutChannel,newOutChannel,e.fn,e.count)()
		} else {
			go NewProcessorRunner(lastProcessorOutChannel,newOutChannel,e.fn)()
		}
		lastProcessorOutChannel = newOutChannel
	}
	return lastProcessorOutChannel
}

func NewProcessorRunner(in fileInfoChannel,out fileInfoChannel,fn processor) func(){
	return func(){
		fn(in, out)
		close(out)
	}
}

func NewProcessorRunnerWithWait(in fileInfoChannel,out fileInfoChannel,fn processor,count int)  func(){
	return func(){
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 1; i <= count; i++ {
			go func() {
				fn(in, out)
				defer wg.Done()
			}()
		}
		wg.Wait()
		close(out)
	}
}