package redis

import (
	"context"
	"time"
)

// ScriptTemplates 常用Lua脚本模板
type ScriptTemplates struct {
	script *Script
}

// NewScriptTemplates 创建脚本模板管理器
func (s *Script) NewScriptTemplates() *ScriptTemplates {
	return &ScriptTemplates{script: s}
}

// RegisterCommonScripts 注册常用脚本
func (st *ScriptTemplates) RegisterCommonScripts(ctx context.Context) error {
	scripts := []*ScriptInfo{
		st.GetDistributedLockScript(),
		st.GetDistributedUnlockScript(),
		st.GetRateLimitScript(),
		st.GetCounterScript(),
		st.GetAtomicIncrementScript(),
		st.GetAtomicDecrementScript(),
		st.GetCompareAndSwapScript(),
		st.GetBatchSetScript(),
		st.GetBatchGetScript(),
		st.GetAtomicListPushScript(),
		st.GetAtomicListPopScript(),
		st.GetAtomicSetAddScript(),
		st.GetAtomicSetRemoveScript(),
		st.GetAtomicHashSetScript(),
		st.GetAtomicHashGetScript(),
		st.GetAtomicZSetAddScript(),
		st.GetAtomicZSetRemoveScript(),
		st.GetAtomicZSetIncrementScript(),
		st.GetAtomicZSetRangeScript(),
		st.GetAtomicZSetRankScript(),
		st.GetAtomicZSetCountScript(),
	}

	for _, script := range scripts {
		if err := st.script.Register(ctx, script); err != nil {
			return err
		}
	}

	return nil
}

// GetDistributedLockScript 分布式锁脚本
func (st *ScriptTemplates) GetDistributedLockScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "distributed_lock",
		Source: `local key = KEYS[1]
local value = ARGV[1]
local ttl = tonumber(ARGV[2])

local result = redis.call('SET', key, value, 'NX', 'EX', ttl)
if result then
    return 1
else
    return 0
end`,
		Keys:        []string{"lock_key"},
		Args:        []string{"lock_value", "ttl_seconds"},
		Description: "分布式锁获取脚本",
		Timeout:     5 * time.Second,
	}
}

// GetDistributedUnlockScript 分布式锁释放脚本
func (st *ScriptTemplates) GetDistributedUnlockScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "distributed_unlock",
		Source: `local key = KEYS[1]
local value = ARGV[1]

if redis.call('GET', key) == value then
    return redis.call('DEL', key)
else
    return 0
end`,
		Keys:        []string{"lock_key"},
		Args:        []string{"lock_value"},
		Description: "分布式锁释放脚本",
		Timeout:     5 * time.Second,
	}
}

// GetRateLimitScript 限流脚本
func (st *ScriptTemplates) GetRateLimitScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "rate_limit",
		Source: `local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call('GET', key)
if current == false then
    redis.call('SET', key, 1, 'EX', window)
    return {1, limit - 1}
end

current = tonumber(current)
if current < limit then
    redis.call('INCR', key)
    return {1, limit - current - 1}
else
    return {0, 0}
end`,
		Keys:        []string{"rate_limit_key"},
		Args:        []string{"limit", "window_seconds"},
		Description: "限流脚本",
		Timeout:     5 * time.Second,
	}
}

// GetCounterScript 计数器脚本
func (st *ScriptTemplates) GetCounterScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "counter",
		Source: `local key = KEYS[1]
local increment = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])

local current = redis.call('GET', key)
if current == false then
    redis.call('SET', key, increment, 'EX', ttl)
    return increment
else
    local new_value = tonumber(current) + increment
    redis.call('SET', key, new_value, 'EX', ttl)
    return new_value
end`,
		Keys:        []string{"counter_key"},
		Args:        []string{"increment", "ttl_seconds"},
		Description: "计数器脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicIncrementScript 原子自增脚本
func (st *ScriptTemplates) GetAtomicIncrementScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_increment",
		Source: `local key = KEYS[1]
local increment = tonumber(ARGV[1])
local max_value = tonumber(ARGV[2])

local current = redis.call('GET', key)
if current == false then
    current = 0
else
    current = tonumber(current)
end

local new_value = current + increment
if max_value > 0 and new_value > max_value then
    return {current, 0}
end

redis.call('SET', key, new_value)
return {new_value, 1}`,
		Keys:        []string{"counter_key"},
		Args:        []string{"increment", "max_value"},
		Description: "原子自增脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicDecrementScript 原子自减脚本
func (st *ScriptTemplates) GetAtomicDecrementScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_decrement",
		Source: `local key = KEYS[1]
local decrement = tonumber(ARGV[1])
local min_value = tonumber(ARGV[2])

local current = redis.call('GET', key)
if current == false then
    current = 0
else
    current = tonumber(current)
end

local new_value = current - decrement
if min_value > 0 and new_value < min_value then
    return {current, 0}
end

redis.call('SET', key, new_value)
return {new_value, 1}`,
		Keys:        []string{"counter_key"},
		Args:        []string{"decrement", "min_value"},
		Description: "原子自减脚本",
		Timeout:     5 * time.Second,
	}
}

// GetCompareAndSwapScript 比较并交换脚本
func (st *ScriptTemplates) GetCompareAndSwapScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "compare_and_swap",
		Source: `local key = KEYS[1]
local expected = ARGV[1]
local new_value = ARGV[2]

local current = redis.call('GET', key)
if current == expected then
    redis.call('SET', key, new_value)
    return {new_value, 1}
else
    return {current, 0}
end`,
		Keys:        []string{"cas_key"},
		Args:        []string{"expected_value", "new_value"},
		Description: "比较并交换脚本",
		Timeout:     5 * time.Second,
	}
}

// GetBatchSetScript 批量设置脚本
func (st *ScriptTemplates) GetBatchSetScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "batch_set",
		Source: `local keys = {}
local values = {}
local ttls = {}

for i = 1, #KEYS do
    table.insert(keys, KEYS[i])
    table.insert(values, ARGV[i])
    table.insert(ttls, tonumber(ARGV[#KEYS + i]))
end

local results = {}
for i = 1, #keys do
    if ttls[i] > 0 then
        redis.call('SET', keys[i], values[i], 'EX', ttls[i])
    else
        redis.call('SET', keys[i], values[i])
    end
    table.insert(results, 1)
end

return results`,
		Keys:        []string{"key1", "key2", "key3"},
		Args:        []string{"value1", "value2", "value3", "ttl1", "ttl2", "ttl3"},
		Description: "批量设置脚本",
		Timeout:     10 * time.Second,
	}
}

// GetBatchGetScript 批量获取脚本
func (st *ScriptTemplates) GetBatchGetScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "batch_get",
		Source: `local results = {}
for i = 1, #KEYS do
    local value = redis.call('GET', KEYS[i])
    table.insert(results, value)
end
return results`,
		Keys:        []string{"key1", "key2", "key3"},
		Args:        []string{},
		Description: "批量获取脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicListPushScript 原子列表推入脚本
func (st *ScriptTemplates) GetAtomicListPushScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_list_push",
		Source: `local key = KEYS[1]
local direction = ARGV[1]
local max_length = tonumber(ARGV[2])

local values = {}
for i = 3, #ARGV do
    table.insert(values, ARGV[i])
end

local current_length = redis.call('LLEN', key)
if max_length > 0 and current_length + #values > max_length then
    return {current_length, 0}
end

local result
if direction == 'left' then
    result = redis.call('LPUSH', key, unpack(values))
else
    result = redis.call('RPUSH', key, unpack(values))
end

return {result, 1}`,
		Keys:        []string{"list_key"},
		Args:        []string{"direction", "max_length", "value1", "value2"},
		Description: "原子列表推入脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicListPopScript 原子列表弹出脚本
func (st *ScriptTemplates) GetAtomicListPopScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_list_pop",
		Source: `local key = KEYS[1]
local direction = ARGV[1]
local count = tonumber(ARGV[2])

local current_length = redis.call('LLEN', key)
if current_length == 0 then
    return {nil, 0}
end

local result
if direction == 'left' then
    if count == 1 then
        result = redis.call('LPOP', key)
    else
        result = redis.call('LPOP', key, count)
    end
else
    if count == 1 then
        result = redis.call('RPOP', key)
    else
        result = redis.call('RPOP', key, count)
    end
end

return {result, 1}`,
		Keys:        []string{"list_key"},
		Args:        []string{"direction", "count"},
		Description: "原子列表弹出脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicSetAddScript 原子集合添加脚本
func (st *ScriptTemplates) GetAtomicSetAddScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_set_add",
		Source: `local key = KEYS[1]
local max_members = tonumber(ARGV[1])

local values = {}
for i = 2, #ARGV do
    table.insert(values, ARGV[i])
end

local current_count = redis.call('SCARD', key)
if max_members > 0 and current_count + #values > max_members then
    return {current_count, 0}
end

local result = redis.call('SADD', key, unpack(values))
return {result, 1}`,
		Keys:        []string{"set_key"},
		Args:        []string{"max_members", "member1", "member2"},
		Description: "原子集合添加脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicSetRemoveScript 原子集合移除脚本
func (st *ScriptTemplates) GetAtomicSetRemoveScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_set_remove",
		Source: `local key = KEYS[1]

local values = {}
for i = 1, #ARGV do
    table.insert(values, ARGV[i])
end

local result = redis.call('SREM', key, unpack(values))
return result`,
		Keys:        []string{"set_key"},
		Args:        []string{"member1", "member2"},
		Description: "原子集合移除脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicHashSetScript 原子哈希设置脚本
func (st *ScriptTemplates) GetAtomicHashSetScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_hash_set",
		Source: `local key = KEYS[1]
local field = ARGV[1]
local value = ARGV[2]
local ttl = tonumber(ARGV[3])

local result = redis.call('HSET', key, field, value)
if ttl > 0 then
    redis.call('EXPIRE', key, ttl)
end
return result`,
		Keys:        []string{"hash_key"},
		Args:        []string{"field", "value", "ttl_seconds"},
		Description: "原子哈希设置脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicHashGetScript 原子哈希获取脚本
func (st *ScriptTemplates) GetAtomicHashGetScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_hash_get",
		Source: `local key = KEYS[1]
local field = ARGV[1]

local value = redis.call('HGET', key, field)
return value`,
		Keys:        []string{"hash_key"},
		Args:        []string{"field"},
		Description: "原子哈希获取脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicZSetAddScript 原子有序集合添加脚本
func (st *ScriptTemplates) GetAtomicZSetAddScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_zset_add",
		Source: `local key = KEYS[1]
local max_members = tonumber(ARGV[1])

local members = {}
for i = 2, #ARGV, 2 do
    table.insert(members, {tonumber(ARGV[i]), ARGV[i + 1]})
end

local current_count = redis.call('ZCARD', key)
if max_members > 0 and current_count + #members > max_members then
    return {current_count, 0}
end

local result = redis.call('ZADD', key, unpack(members))
return {result, 1}`,
		Keys:        []string{"zset_key"},
		Args:        []string{"max_members", "score1", "member1", "score2", "member2"},
		Description: "原子有序集合添加脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicZSetRemoveScript 原子有序集合移除脚本
func (st *ScriptTemplates) GetAtomicZSetRemoveScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_zset_remove",
		Source: `local key = KEYS[1]

local members = {}
for i = 1, #ARGV do
    table.insert(members, ARGV[i])
end

local result = redis.call('ZREM', key, unpack(members))
return result`,
		Keys:        []string{"zset_key"},
		Args:        []string{"member1", "member2"},
		Description: "原子有序集合移除脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicZSetIncrementScript 原子有序集合自增脚本
func (st *ScriptTemplates) GetAtomicZSetIncrementScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_zset_increment",
		Source: `local key = KEYS[1]
local member = ARGV[1]
local increment = tonumber(ARGV[2])

local result = redis.call('ZINCRBY', key, increment, member)
return result`,
		Keys:        []string{"zset_key"},
		Args:        []string{"member", "increment"},
		Description: "原子有序集合自增脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicZSetRangeScript 原子有序集合范围查询脚本
func (st *ScriptTemplates) GetAtomicZSetRangeScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_zset_range",
		Source: `local key = KEYS[1]
local start = tonumber(ARGV[1])
local stop = tonumber(ARGV[2])
local with_scores = ARGV[3] == 'true'

local result
if with_scores then
    result = redis.call('ZRANGE', key, start, stop, 'WITHSCORES')
else
    result = redis.call('ZRANGE', key, start, stop)
end

return result`,
		Keys:        []string{"zset_key"},
		Args:        []string{"start", "stop", "with_scores"},
		Description: "原子有序集合范围查询脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicZSetRankScript 原子有序集合排名脚本
func (st *ScriptTemplates) GetAtomicZSetRankScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_zset_rank",
		Source: `local key = KEYS[1]
local member = ARGV[1]
local reverse = ARGV[2] == 'true'

local result
if reverse then
    result = redis.call('ZREVRANK', key, member)
else
    result = redis.call('ZRANK', key, member)
end

return result`,
		Keys:        []string{"zset_key"},
		Args:        []string{"member", "reverse"},
		Description: "原子有序集合排名脚本",
		Timeout:     5 * time.Second,
	}
}

// GetAtomicZSetCountScript 原子有序集合计数脚本
func (st *ScriptTemplates) GetAtomicZSetCountScript() *ScriptInfo {
	return &ScriptInfo{
		Name: "atomic_zset_count",
		Source: `local key = KEYS[1]
local min_score = ARGV[1]
local max_score = ARGV[2]

local result = redis.call('ZCOUNT', key, min_score, max_score)
return result`,
		Keys:        []string{"zset_key"},
		Args:        []string{"min_score", "max_score"},
		Description: "原子有序集合计数脚本",
		Timeout:     5 * time.Second,
	}
}
