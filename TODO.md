# TODO

## После фронтенда

### Бэкенд
- [ ] Продление аренды — `POST /rentals/{id}/extend`
- [ ] Пагинация в `List`-методах
- [ ] Фильтры — аренды за период, поломки по статусу, велосипеды по availability
- [ ] Payouts для партнёра — экспорт PDF/CSV
- [ ] Тесты — unit для usecase, integration для repository
- [ ] Обновить README.md
- [ ] Seed — флаг «не перезаписывать пароль»
- [ ] Refresh-токены
- [ ] Смена пароля + инвалидация токенов

### Инфраструктура
- [ ] Dockerfile для API
- [ ] Docker-compose с API + Postgres + мигратор
- [ ] CI/CD — GitHub Actions
- [ ] Деплой на VPS/облако