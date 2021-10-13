-- delete reports without email
delete from bug_reports where email is null or email = '';

-- Update schnelltestportal bug receivers
update operators set bug_reports_receiver = 'center' where operator_number = 'cwa-schnelltestportal';