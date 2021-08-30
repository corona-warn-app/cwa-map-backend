update centers
set dcc = true
where operator_uuid in (select uuid
                        from operators
                        where operator_number in ('A1026', 'A1024', 'A1042', 'A1005', 'A1011'));