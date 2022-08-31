alter table operators
    add column notified timestamp;

alter table operators
    add column notification_token varchar;

alter table centers
    add column notified timestamp;

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
[LINK]<br>
<br>
Viele Grüße<br>
Ihr Corona-Warn-App Team
</body>
</html>');
