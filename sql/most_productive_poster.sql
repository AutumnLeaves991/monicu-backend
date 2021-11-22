select u.discord_id, count(distinct p.id) as posts
from "user" u
         join post p on u.id = p.user_id
group by u.id
order by posts desc