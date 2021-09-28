create table system_settings
(
    config_key   varchar not null primary key,
    config_value text
);

INSERT INTO public.system_settings (config_key, config_value) VALUES ('reports.email.default', 'dirk.reske@t-systems.com');
INSERT INTO public.system_settings (config_key, config_value) VALUES ('reports.email.subject', 'Meldungen zur Ihren Teststellen');
INSERT INTO public.system_settings (config_key, config_value) VALUES ('reports.email.template', 'Guten Tag,<br>
<br>
für Ihre Teststellen wurden folgende Fehler gemeldet.<br>
<br>
{{range $center, $reports := .Centers}}Meldungen für die Teststelle {{(index $reports 0).CenterName}} ({{(index $reports 0).CenterAddress}})<br>
<ul>
{{range $reportUUID, $report := $reports}}
<li>{{$report.Subject}} {{if $report.Message}}(Hinweis: {{$report.Message}}){{end}} (Gemeldet am {{$report.Created}}){{end}}</li>
</ul>
<br>
{{end}}
<br>
Viele Grüße<br>
Ihr Schnelltestportal.de Team');