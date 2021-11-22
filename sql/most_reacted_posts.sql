select p.*, count(distinct ur.user_id) as reactions
from post p
         join reaction r on p.id = r.post_id
         join user_reaction ur on r.id = ur.reaction_id
group by p.id
order by reactions desc