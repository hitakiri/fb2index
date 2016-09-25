// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package cache is a simple expiring cache.
package cache

import (
	"runtime"
	"sync"
	"time"
)

type Cache struct {
	m     sync.RWMutex
	items map[string]cacheItem

	expire     time.Duration
	gcInterval time.Duration
}

type cacheItem struct {
	expire time.Time
	data   []byte
}

func New(expire, gcInterval time.Duration) *Cache {
	c := Cache{
		items:      make(map[string]cacheItem),
		expire:     expire,
		gcInterval: gcInterval,
	}
	go c.gc()
	return &c
}

func (c *Cache) gc() {
	tick := time.NewTicker(c.gcInterval)
	runtime.SetFinalizer(c, func(*Cache) { tick.Stop() })
	for range tick.C {
		now := time.Now()
		c.m.Lock()
		for k, v := range c.items {
			if v.expire.Before(now) {
				delete(c.items, k)
			}
		}
		c.m.Unlock()
	}
}

func (c *Cache) Get(k string) []byte {
	c.m.RLock()
	item, ok := c.items[k]
	c.m.RUnlock()
	if !ok {
		return nil
	}
	return item.data
}

func (c *Cache) Put(k string, v []byte) {
	c.m.Lock()
	c.items[k] = cacheItem{
		expire: time.Now().Add(c.expire),
		data:   v,
	}
	c.m.Unlock()
}
