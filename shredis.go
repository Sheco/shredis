package shredis

import (
    "errors"
    "strconv"
    "container/list"
    "time"
    "sync"
)

type Shredis struct {
    memory map[string]string
    hmemory map[string]map[string]string
    qmemory map[string]*list.List
    mutex sync.Mutex

}

func New() *Shredis {
    var this Shredis

    this.memory = make(map[string]string)
    this.hmemory = make(map[string]map[string]string)
    this.qmemory = make(map[string]*list.List)

    return &this
}
func arrayToMap(values []interface{}) (retMap map[string]string) {
	retMap = make(map[string]string)
	for i:=0; i< len(values); i+=2 {
		retMap[values[i].(string)] = values[i+1].(string)
	}
	return
}

func (this *Shredis) Hset(key string, values ...interface{}) error {
    this.mutex.Lock()
    defer this.mutex.Unlock()

	var dataMap map[string]string
	if len(values) > 1 {
		dataMap = arrayToMap(values)	
	} else if len(values) == 1 {
		dataMap = values[0].(map[string]string)
	} else {
	    return errors.New("Invalid arguments")
	}

    // if key does not exists, a new key is created
    if _, ok := this.hmemory[key]; !ok {
        this.hmemory[key] = make(map[string]string)
    }

    for field, value := range dataMap {
        this.hmemory[key][field] = value
    }
    return nil
}

func (this *Shredis) Hincrby(key string, field string, amount int) (string, error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    // if key does not exists, a new key is created
    if _, ok := this.hmemory[key]; !ok {
        this.hmemory[key] = make(map[string]string)
    }

    // if value does not exists, set it to 0
    if _, ok := this.hmemory[key][field]; !ok {
        this.hmemory[key][field] = "0"
    }

    value, err := strconv.Atoi(this.hmemory[key][field])
    if (err != nil) {
        return "0", errors.New("Field is not numeric")
    }
    
    retvalue := strconv.Itoa(value+amount)
    this.hmemory[key][field] = retvalue

    return this.hmemory[key][field], nil
}

func (this *Shredis) Hget(key string, field string) (string, error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if _, ok := this.hmemory[key]; !ok {
        return "", errors.New("Invalid key")
    }

    if _, ok := this.hmemory[key][field]; !ok {
        return "", errors.New("Invalid field")
    }

    return this.hmemory[key][field], nil
}

func (this *Shredis) Hgetall(key string) (map[string]string, error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    // if key does not exists, a new key is created
    if _, ok := this.hmemory[key]; !ok {
        this.hmemory[key] = make(map[string]string)
    }

    return this.hmemory[key], nil
}

func (this *Shredis) Hexists(key string, field string) (bool, error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if _, ok := this.hmemory[key]; !ok {
        return false, nil
    }
    if _, ok := this.hmemory[key][field]; !ok {
        return false, nil
    }

    return true, nil
}

func (this *Shredis) Set(key string, value string) (error) {
    this.Hset("|s|"+key, "default", value)
    return nil
}

func (this *Shredis) Get(key string) (string, error) {
    return this.Hget("|s|"+key, "default")
}

func (this *Shredis) Lpush(key string, value string)  (error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    // if key does not exists, a new key is created
    if _, ok := this.qmemory[key]; !ok {
        this.qmemory[key] = list.New()
    }

    this.qmemory[key].PushFront(value)
    return nil
}

func (this *Shredis) Rpush(key string, value string) (error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    // if key does not exists, a new key is created
    if _, ok := this.qmemory[key]; !ok {
        this.qmemory[key] = list.New()
    }

    this.qmemory[key].PushBack(value)
    return nil
}

func (this *Shredis) Lpop(key string)  (string, error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    // if key does not exists, retur nil
    if _, ok := this.qmemory[key]; !ok {
        return "", errors.New("Invalid list")
    }

    element := this.qmemory[key].Front()
    if element == nil {
        return "", errors.New("Empty list")
    }

    this.qmemory[key].Remove(element)

    return element.Value.(string), nil
}

func (this *Shredis) Rpop(key string)  (string, error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    // if key does not exists, retur nil
    if _, ok := this.qmemory[key]; !ok {
        return "", errors.New("Invalid list")
    }

    element := this.qmemory[key].Back()
    if element == nil {
        return "", errors.New("Empty list")
    }
    this.qmemory[key].Remove(element)

    return element.Value.(string), nil
}

func (this *Shredis) Blpop(key string, timeout int) (string, error) {
    for {
        value, err := this.Lpop(key)
        if err == nil {
            return value, nil
        }
        time.Sleep(1 * time.Second)
        timeout--
        if timeout <= 0 {
            return "", errors.New("Timeout")
        }
    }
}

func (this *Shredis) Brpop(key string, timeout int) (string, error) {
    for {
        value, err := this.Rpop(key)
        if err == nil {
            return value, nil
        }
        time.Sleep(1 * time.Second)
        timeout--
        if timeout <= 0 {
            return "", errors.New("Timeout")
        }
    }
}


