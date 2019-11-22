#### docker 启动arango

docker启动:https://hub.docker.com/_/arangodb/

```
docker run -e ARANGO_RANDOM_ROOT_PASSWORD=1 -d arangodb
//随机生成一个密码.默认用户名root,,docker logs 查看密码
```

SQL|AQL
---|---
database | database
table| collection
row| document
column| attribute
table joins| collection joins
primary key| _key
index| index


场景|SQL|AQL
----|--------------|---
插入一条|INSERT INTO users (name, gender)  VALUES ("John Doe", "m");|INSERT {name: "John Doe", gender: "m" } INTO users
插入多条|INSERT INTO users (name, gender)   VALUES ("John Doe", "m"),("Jane Smith", "f")|FOR user IN [{ name: "John Doe", gender: "m" }, { name: "Jane Smith", gender: "f" } ] INSERT user INTO users
更新|UPDATE users SET name = "John Smith" WHERE id = 1;|UPDATE { _key:"1" }  WITH { name: "John Smith" }  IN users
删除id|DELETE FROM users  WHERE id = 1;|REMOVE { _key:"1" } IN users
删除多个|DELETE FROM users WHERE active = 1|FOR user IN users  FILTER user.active == 1  REMOVE user IN users
查询|selct * from users | for u in users return u
排序|SELECT * FROM users WHERE active = 1 ORDER BY name, gender| FOR user IN users  FILTER user.active == 1  SORT user.name, user.gender  RETURN user
条数|SELECT gender, COUNT(*) AS number FROM users  WHERE active = 1  GROUP BY gender|FOR user IN users  FILTER user.active == 1  COLLECT gender = user.gender    WITH COUNT INTO number  RETURN {     gender: gender,    number: number   }
分组|SELECT YEAR(dateRegister) AS year,       MONTH(dateRegister) AS month,  COUNT(*) AS number  FROM users WHERE active = 1  GROUP BY year, month HAVING number > 20;|FOR user IN users  FILTER user.active == 1  COLLECT    year = DATE_YEAR(user.dateRegistered),     month = DATE_MONTH(user.dateRegistered)    WITH COUNT INTO number   FILTER number > 20    RETURN {      year: year,      month: month,      number: number    }
最大值最小值|SELECT MIN(dateRegistered) AS minDate,  MAX(dateRegistered) AS maxDate  FROM users WHERE active = 1;|FOR user IN users FILTER user.active == 1 COLLECT AGGREGATE minDate = MIN(user.dateRegistered),  maxDate = MAX(user.dateRegistered) RETURN { minDate, maxDate }
join|SELECT * FROM users INNER JOIN friends  ON (friends.user = users.id);|FOR user IN users  FOR friend IN friends  FILTER friend.user == user._key  RETURN MERGE(user, friend)//合并
left join|SELECT * FROM users LEFT JOIN friends ON (friends.user = users.id); |FOR user IN users  LET friends = ( FOR friend IN friends   FILTER friend.user == user._key     RETURN friend )  FOR friendToJoin IN (  LENGTH(friends) > 0 ? friends : [ { /* no match exists */ } ]  ) RETURN {    user: user,  friend: friend  }

## graphs-->aql

1.一种是找层级(几度好友)    
2.一种是找最近的路径    

1.语法:

FOR vertex[, edge[, path]]   
  IN [min[..max]]   
  OUTBOUND|INBOUND|ANY startVertex   
  GRAPH graphName   
  [OPTIONS options]   

vertex:点   
edge:线  
path:点和线组成(分:vertices和edges)    
min..max:从几度到第几度    
OUTBOUND|INBOUND|ANY:从某个顶点传出.传出或双向    
GRAPH graphName: graphs表名称    
 
例如:查询某个用户1度好友关系
```
    for v,e in 1 OUTBOUND @user_key GRAPH 'user_follow'
			filter e._to!=@user_key
			return v.user_id

```
@user_key等以@开头的为占位符,防止注入
```
语法:@分割
FOR u IN users
  FILTER u.id == @id && u.name == @name
  RETURN u
api:
{
  "query": "FOR u IN users FILTER u.id == @id && u.name == @name RETURN u",
  "bindVars": {
    "id": 123,
    "name": "John Smith"
  }
}
```

2.语法:   

FOR vertex[, edge]   
  IN OUTBOUND|INBOUND|ANY SHORTEST_PATH   
  startVertex TO targetVertex   
  GRAPH graphName   
  [OPTIONS options]   
例如:
```
FOR v, e IN OUTBOUND SHORTEST_PATH 'circles/A' 
TO 'circles/D' 
GRAPH 'traversalGraph'
RETURN [v._key, e._key]

```

AQL:函数:https://docs.arangodb.com/3.2/AQL/Functions/   
Aql操作:https://docs.arangodb.com/3.2/AQL/Operations/
