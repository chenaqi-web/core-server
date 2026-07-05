package scripts

const (
	ThumbUpLuaScript = `
local keyZset = KEYS[1]
local keyCount = KEYS[2]
local max_len = tonumber(ARGV[1])
local score = tonumber(ARGV[2])
local object_id = ARGV[3]
local ttl_like = tonumber(ARGV[4])
local ttl_count = tonumber(ARGV[5])

-- 1. 判断 object_id 是否已存在 0表示存在
local exists_score  = redis.call("ZSCORE", keyZset, object_id)
if exists_score  then
	return {
		0,
		"already exists"
	}
end

-- 2. 不存在或者需要更新，执行添加
redis.call("ZADD", keyZset, score, object_id)
redis.call("INCR", keyCount)

-- 3. 检查长度是否超过最大限制
local current_len = redis.call('ZCARD',KEYS[1])
if current_len > max_len then
    -- 需要删除最旧的 N 条（score 最小）
    local remove_count = current_len - max_len

    -- 按 rank 删除：0 ~ remove_count-1 (超过就截取)
    redis.call("ZREMRANGEBYRANK", keyZset, 0, remove_count - 1)
end

-- 4.设置过期时间（刷新 TTL，count可能会有不一致的问题）
redis.call("EXPIRE", keyZset, ttl_like)
redis.call("EXPIRE", keyCount, ttl_count)


return {
    1,
    "success"
}
`
	CancelThumbUpLuaScript = `
local zset_key = KEYS[1]
local count_key = KEYS[2]
local member = ARGV[1]

-- 1. 获取并判断点赞关系是否存在，同时保存 score 以便后续返回
local score = redis.call("ZSCORE", zset_key, member)
if not score then
    return {
        0,
        "not_found"
    }
end

-- 2. 删除点赞关系
redis.call("ZREM", zset_key, member)

-- 3. 点赞数 -1（点赞数>0才可以删除）
local count = redis.call("GET", count_key)
if tonumber(count) > 0 then
    redis.call("DECR", count_key)
end

-- 成功，返回被删除元素的 score
return {
    1,
    score
}
`
)
