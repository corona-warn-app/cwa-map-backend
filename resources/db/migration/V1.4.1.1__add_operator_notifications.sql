alter table operators
    add column notified timestamp;

alter table operators
    add column notification_token varchar;

alter table centers
    add column notified timestamp;

update operators set bug_reports_receiver = 'operator' where bug_reports_receiver is null;
update centers set last_update = '2022-03-01 00:00:00.000000 +00:00' where last_update is null;

INSERT INTO public.system_settings (config_key, config_value)
VALUES ('operator.notification.subject', 'Meldungen zur Ihren Teststellen');
INSERT INTO public.system_settings (config_key, config_value)
VALUES ('operator.notification.template', '<html>
<head>
    <title>Veraltete Teststellen</title>
    <meta charset="UTF-8">
</head>
<body>
Sehr geehrte Damen und Herren,<br>
<br>
wir haben festgestellt, dass keine Ihrer Teststellen innerhalb der letzten vier Wochen aktualisiert wurde. Dies bedeutet, dass den Nutzern ab sofort eine Warnung angezeigt wird, dass wir keine Aktualisierungen des Betreibers mehr erhalten. Um diese Warnung wieder zu entfernen, bitten wir Sie, Ihre Teststellen regelmäßig zu aktualisieren, auch dann, wenn es keine Änderungen der Daten gibt.<br>
<br>
Zusätzlich bitten wir Sie, den folgenden Link zur Bestätigung Ihrer E-Mail Adresse zu klicken.<br>
<a href="https://map-admin.schnelltestportal.de/api/operators/confirm/{{.NotificationToken}}">Bestätigen Sie Ihre E-Mail Adresse</a><br>
<br>
Viele Grüße<br>
Ihr Corona-Warn-App Team
</body>
</html>');

INSERT INTO public.system_settings (config_key, config_value)
VALUES ('center.notification.subject', 'Meldungen zur Ihren Teststellen');
INSERT INTO public.system_settings (config_key, config_value)
VALUES ('center.notification.template', '<html>
<head>
    <title>Veraltete Teststellen</title>
    <meta charset="UTF-8">
</head>
<body>
Sehr geehrte Damen und Herren,<br>
<br>
wir haben festgestellt, dass Ihre Teststelle ''{{.Name}}'' seit über vier Wochen nicht aktualisiert wurde. Dies bedeutet, dass den Nutzern ab sofort eine Warnung angezeigt wird, dass wir keine Aktualisierungen des Betreibers mehr erhalten. Um diese Warnung wieder zu entfernen, bitten wir Sie, Ihre Teststellen regelmäßig zu aktualisieren, auch dann, wenn es keine Änderungen der Daten gibt.<br>
<br>
Viele Grüße<br>
Ihr Corona-Warn-App Team
</body>
</html>');
