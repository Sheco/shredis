package shredis

import (
    "testing"
    "sync"
    "strconv"
)

func TestHset(t *testing.T) {
    redis := New()
    redis.Hset("hset", "field", "value")

    value, _ := redis.Hget("hset", "field")
    if value != "value" {
        t.Errorf("hset/hget failed")
    }

    redis.Hset("hmset", "field1", "value1", "field2", "value2")
    value, _ = redis.Hget("hmset", "field1")
    if value != "value1" {
        t.Errorf("hset failed with variadic arguments")
    }

    redis.Hset("hmset2", map[string]string {
        "FIELD1": "VALUE1",
        "FIELD2": "VALUE2",
    })
    value, _ = redis.Hget("hmset2", "FIELD1")
    if value != "VALUE1" {
        t.Errorf("hset failed with a map[string]string")
    }
}

func TestHincrby(t *testing.T) {
    redis := New()
    value, _ := redis.Hincrby("hincrby", "field", 1)
    if value != "1" {
        t.Error("First hincrby test failed")
    }

    redis.Hincrby("hincrby", "field", 1)
    value, _ = redis.Hget("hincrby", "field")

    if value != "2" {
        t.Error("Second hincrby test failed")
    }
}
    
func TestHexists(t *testing.T) {
    redis := New()
    exists, _ := redis.Hexists("invalid", "invalid")

    if (exists) {
        t.Error("hexists invalid/invalid failed")
    }

    exists, _ = redis.Hexists("hset", "invalid")
    if (exists) {
        t.Error("hexists valid/invalid failed")
    }

    redis.Hset("hset", "field", "value")
    exists, _ = redis.Hexists("hset", "field")
    if (!exists) {
        t.Error("hexists valid/valid failed")
    }
}

func TestQueue(t *testing.T) {
    redis := New()
    redis.Rpush("queue", "1")
    redis.Rpush("queue", "2")
    redis.Rpush("queue", "3")
    redis.Rpush("queue", "4")

    value, err := redis.Rpop("queue")

    if value != "4" {
        t.Error("rpop failed")
    }
    value, _ = redis.Lpop("queue")
    if value != "1" {
        t.Error("lpop failed")
    }
    value, _ = redis.Rpop("queue")
    value, _ = redis.Rpop("queue")
    value, err = redis.Rpop("queue")
    if (err == nil) {
        t.Error("Rpop expected an error, because the queue should be empty")
    }
}

func TestBlpop(t *testing.T) {
    redis := New()
    redis.Rpush("queue", "1")
    
    _, err := redis.Blpop("queue", 1)
    if err != nil {
        t.Errorf("A queue with items returned: %s", err)
    }

    _, err = redis.Blpop("queue", 1)
    if err == nil {
        t.Errorf("An empty queue didn't return a timeout error")
    }

}


func TestConcurrency(t *testing.T) {
    redis := New()
    wait := sync.WaitGroup{}

    increments := 1000
    producers := 1000

    producer := func() {
        defer wait.Done()
        for i:= 0; i <increments ; i++ {
            redis.Hincrby("test", "counter", 1)
        }
    }

    for i := 0; i < producers; i++ {
        wait.Add(1)
        go producer()
    }
    wait.Wait()

    value, _ := redis.Hget("test", "counter")
    t.Logf("Concurrency result: (%s)\n", value)

    if value != strconv.Itoa(increments*producers) {
        t.Error("Concurrency failed")
    }
}

