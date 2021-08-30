delete
from centers a using centers b
where (a.uuid > b.uuid and a.name = b.name and a.address = b.address)
   or (a.name = '');