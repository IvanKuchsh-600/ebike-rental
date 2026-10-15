

document.addEventListener('DOMContentLoaded', renderNav);

// ============ Token management ============
function getToken() {
    return localStorage.getItem('token');
}

function getUserRole() {
    const token = getToken();
    if (!token) return null;

    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.role;
    } catch (err) {
        console.error('Failed to parse token:', err);
        return null;
    }
}

function setToken(token) {
    localStorage.setItem('token', token);
}

function clearToken() {
    localStorage.removeItem('token');
}

// ============ Login ============

async function handleLogin(event) {
    event.preventDefault();

    const form = event.target;
    const email = form.email.value;
    const password = form.password.value;
    const errorDiv = document.getElementById('error');
    errorDiv.textContent = '';

    try {
        const response = await fetch('/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        if (!response.ok) {
            const data = await response.json();
            errorDiv.textContent = data.error || 'Ошибка входа';
            return;
        }

        const data = await response.json();
        setToken(data.token);

        // Редирект в зависимости от роли
        if (data.user.role === 'owner') {
            window.location.href = '/web/dashboard';
        } else if (data.user.role === 'partner') {
            window.location.href = '/web/partner/dashboard';
        } else {
            errorDiv.textContent = 'Неизвестная роль';
        }
    } catch (err) {
        errorDiv.textContent = 'Ошибка сети: ' + err.message;
    }
}

// ============ Logout ============

function logout() {
    clearToken();
    window.location.href = '/web/login';
}

// ============ Check auth on page load ============

function requireAuth() {
    if (!getToken()) {
        window.location.href = '/web/login';
    }
}

// ============ Dashboard ============

async function loadDashboard() {
    // 1. Проверяем токен
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    // 2. Проверяем роль — если не owner, редирект на партнёрскую страницу
    try {
        const userResp = await fetch('/auth/me', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (userResp.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (userResp.ok) {
            const user = await userResp.json();
            if (user.role !== 'owner') {
                window.location.href = '/web/partner/dashboard';
                return;
            }
        }
    } catch (err) {
        console.error('Failed to check role:', err);
        return;
    }

    // 3. Пользователь — owner. Загружаем данные
    loadUserInfo(token);
    loadBikes(token);
}

async function loadUserInfo(token) {
    try {
        const response = await fetch('/auth/me', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            // Токен невалиден — редирект на логин
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) return;

        const user = await response.json();
        const userInfo = document.getElementById('user-info');
        if (userInfo) {
            userInfo.textContent = `${user.email} (${user.role})`;
        }
    } catch (err) {
        console.error('Failed to load user info:', err);
    }
}

async function loadBikes(token, showActions = false) {
    const loading = document.getElementById('bikes-loading');
    const errorDiv = document.getElementById('bikes-error');
    const table = document.getElementById('bikes-table');
    const tbody = document.getElementById('bikes-tbody');
    const empty = document.getElementById('bikes-empty');

    try {
        const response = await fetch('/bikes', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }
        if (response.status === 403) {
            window.location.href = '/web/partner/dashboard';
            return;
        }

        if (!response.ok) {
            throw new Error('Ошибка загрузки: ' + response.status);
        }

        const bikes = await response.json();

        // Обновляем статистику (только на dashboard)
        updateStats(bikes);

        if (bikes.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

        empty.style.display = 'none';    // ← СКРЫВАЕМ пустое состояние
        table.style.display = 'table';
        tbody.innerHTML = '';

        for (const bike of bikes) {
            const row = document.createElement('tr');

            let statusBadge = '';
            if (bike.is_broken) {
                statusBadge = '<span style="color: #c0392b; font-weight: 600;">Сломан</span>';
            } else if (bike.is_rented) {
                statusBadge = '<span style="color: #e67e22; font-weight: 600;">Сдан</span>';
            } else {
                statusBadge = '<span style="color: #27ae60; font-weight: 600;">Свободен</span>';
            }

            // Кнопки действий — только если showActions = true
            let actionsCell = '';
            if (showActions) {
                actionsCell = `
                    <td>
                        <div class="actions-cell">
                            <button class="outline btn-sm" onclick="showBikeDetails(${bike.id}, '${escapeHtml(bike.serial_number)}')">Подробнее</button>
                            <button class="outline btn-sm" onclick="openEditBikeForm(${bike.id})">Изменить</button>
                            <button class="outline secondary btn-icon" onclick="deleteBike(${bike.id})" title="Удалить">🗑</button>
                        </div>
                    </td>
                `;
            }

            // Формируем фото-ячейку
            let photoCell = '<span style="color: gray;">—</span>';
            if (bike.photo_url) {
                photoCell = `<img src="${bike.photo_url}" alt="Фото" style="width: 50px; height: 50px; object-fit: cover; border-radius: 4px;">`;
            }

            row.innerHTML = `
                <td>${photoCell}</td>
                <td>${bike.id}</td>
                <td>${escapeHtml(bike.serial_number)}</td>
                <td>${escapeHtml(bike.model || '—')}</td>
                <td>${statusBadge}</td>
                ${actionsCell}
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load bikes:', err);
    }
}

function updateStats(bikes) {
    // Проверяем — есть ли элементы статистики (они только на dashboard)
    const statTotal = document.getElementById('stat-total');
    if (!statTotal) {
        // Не на dashboard — пропускаем
        return;
    }

    let rented = 0;
    let broken = 0;
    let available = 0;

    for (const bike of bikes) {
        if (bike.is_broken) {
            broken++;
        } else if (bike.is_rented) {
            rented++;
        } else {
            available++;
        }
    }

    statTotal.textContent = bikes.length;
    document.getElementById('stat-rented').textContent = rented;
    document.getElementById('stat-broken').textContent = broken;
    document.getElementById('stat-available').textContent = available;
}

// ============ Partner Dashboard ============

async function loadPartnerDashboard() {
    // 1. Проверяем токен
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    // 2. Проверяем роль — если не partner, редирект на owner-страницу
    try {
        const userResp = await fetch('/auth/me', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (userResp.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (userResp.ok) {
            const user = await userResp.json();
            if (user.role !== 'partner') {
                window.location.href = '/web/dashboard';
                return;
            }
        }
    } catch (err) {
        console.error('Failed to check role:', err);
        return;
    }

    // 3. Пользователь — partner. Загружаем данные
    loadUserInfo(token);

    try {
        const response = await fetch('/partner/dashboard', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (response.status === 403) {
            // На всякий случай — если API вернул 403
            window.location.href = '/web/dashboard';
            return;
        }

        if (!response.ok) {
            throw new Error('Ошибка загрузки: ' + response.status);
        }

        const data = await response.json();

        // Статистика по велосипедам
        document.getElementById('stat-total').textContent = data.bikes.total;
        document.getElementById('stat-rented').textContent = data.bikes.rented;
        document.getElementById('stat-broken').textContent = data.bikes.broken;
        document.getElementById('stat-available').textContent = data.bikes.available;

        // Финансы
        document.getElementById('month-label').textContent = data.current_month.month;
        document.getElementById('stat-revenue').textContent = formatMoney(data.current_month.revenue);
        document.getElementById('stat-partner-share').textContent = formatMoney(data.current_month.partner_share);
    } catch (err) {
        console.error('Failed to load partner dashboard:', err);
        alert('Ошибка загрузки данных: ' + err.message);
    }
}


function formatMoney(amount) {
    // Форматируем как "45 000 ₽"
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: 'RUB',
        maximumFractionDigits: 0,
    }).format(amount);
}

// ============ Navigation ============

function renderNav() {
    const navMenu = document.getElementById('nav-menu');
    if (!navMenu) return;

    const token = getToken();
    const currentPath = window.location.pathname;

    // Функция для формирования ссылки с активным состоянием
    const link = (href, label) => {
        const isActive = currentPath === href;
        return `<li><a href="${href}"${isActive ? ' class="active"' : ''}>${label}</a></li>`;
    };

    // Если нет токена
    if (!token) {
        navMenu.innerHTML = link('/web/login', 'Войти');
        return;
    }

    const role = getUserRole();

    let menuItems = '';

    if (role === 'owner') {
        menuItems = [
            link('/web/dashboard', 'Дашборд'),
            link('/web/bikes', 'Велосипеды'),
            link('/web/rentals', 'Аренды'),
            link('/web/breakdowns', 'Поломки'),
            link('/web/payouts', 'Выплаты'),
            `<li><a href="#" onclick="logout(); return false;">Выйти</a></li>`,
        ].join('');
    } else if (role === 'partner') {
        menuItems = [
            link('/web/partner/dashboard', 'Дашборд'),
            link('/web/partner/bikes', 'Велосипеды'),
            link('/web/partner/rentals', 'Аренды'),
            link('/web/partner/breakdowns', 'Поломки'),
            link('/web/partner/payouts', 'Выплаты'),
            `<li><a href="#" onclick="logout(); return false;">Выйти</a></li>`,
        ].join('');
    } else {
        menuItems = link('/web/login', 'Войти');
    }

    navMenu.innerHTML = menuItems;
}

// ============ Auto-init ============

document.addEventListener('DOMContentLoaded', renderNav);

// ============ Bikes page (owner) ============

async function loadBikesPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    // Проверка роли — только для owner
    const user = await checkRole(token, 'owner');
    if (!user) return;

    loadUserInfo(token);
    loadBikes(token, true);
}

async function deleteBike(id) {
    if (!confirm('Удалить велосипед #' + id + '?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch('/bikes/' + id, {
            method: 'DELETE',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (response.status === 409) {
            alert('Этот велосипед нельзя удалить: он был в аренде.\nИстория аренд должна сохраниться.');
            return;
        }

        if (!response.ok && response.status !== 204) {
            const err = await response.json();
            alert('Ошибка удаления: ' + (err.error || response.status));
            return;
        }

        loadBikes(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

// ============ Helper: checkRole ============

async function checkRole(token, expectedRole) {
    try {
        const resp = await fetch('/auth/me', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (resp.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return null;
        }

        if (!resp.ok) return null;

        const user = await resp.json();
        if (user.role !== expectedRole) {
            // Редирект в зависимости от роли
            if (user.role === 'partner') {
                window.location.href = '/web/partner/dashboard';
            } else {
                window.location.href = '/web/dashboard';
            }
            return null;
        }

        return user;
    } catch (err) {
        console.error('Failed to check role:', err);
        return null;
    }
}

// ============ Rentals page (owner) ============

async function loadRentalsPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    const user = await checkRole(token, 'owner');
    if (!user) return;

    loadUserInfo(token);
    loadRentals(token);
}

async function loadRentals(token) {
    const loading = document.getElementById('rentals-loading');
    const errorDiv = document.getElementById('rentals-error');
    const table = document.getElementById('rentals-table');
    const tbody = document.getElementById('rentals-tbody');
    const empty = document.getElementById('rentals-empty');

    try {
        // Загружаем аренды и велосипеды (для серийников)
        const [rentalsResp, bikesResp] = await Promise.all([
            fetch('/rentals', { headers: { 'Authorization': 'Bearer ' + token } }),
            fetch('/bikes', { headers: { 'Authorization': 'Bearer ' + token } }),
        ]);

        loading.style.display = 'none';

        if (rentalsResp.status === 401 || bikesResp.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!rentalsResp.ok || !bikesResp.ok) {
            throw new Error('Ошибка загрузки данных');
        }

        const rentals = await rentalsResp.json();
        const bikes = await bikesResp.json();

        // Map: bike_id → serial_number
        const serialByBikeId = {};
        for (const bike of bikes) {
            serialByBikeId[bike.id] = bike.serial_number;
        }

        if (rentals.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

// Если дошли сюда — значит аренды есть
        empty.style.display = 'none';   // ← ДОБАВИТЬ
        table.style.display = 'table';
        tbody.innerHTML = '';

        for (const rental of rentals) {
            const row = document.createElement('tr');

            const serial = serialByBikeId[rental.bike_id] || '#' + rental.bike_id;

            // Статус аренды
            let statusBadge;
            if (rental.returned_at) {
                statusBadge = '<span style="color: green;">Возвращён</span>';
            } else {
                // Проверяем, просрочена ли
                const plannedEnd = new Date(rental.planned_end_at);
                if (plannedEnd < new Date()) {
                    statusBadge = '<span style="color: red;">Просрочен</span>';
                } else {
                    statusBadge = '<span style="color: orange;">Активен</span>';
                }
            }

            // Статус оплаты
            const paidBadge = rental.is_paid
                ? '<span style="color: green;">Оплачено</span>'
                : '<span style="color: red;">Не оплачено</span>';

            // Кнопка действия
            let actionButton = '';
            if (!rental.returned_at) {
                actionButton = `<button class="outline btn-sm" onclick="openEditRentalForm(${rental.id})">Редактировать</button> `;
                actionButton += `<button class="outline btn-sm" onclick="returnRental(${rental.id})">Вернуть</button> `;
            }
            if (!rental.is_paid && !rental.returned_at) {
                actionButton += `<button class="outline btn-sm" onclick="markRentalPaid(${rental.id})">Оплачено</button> `;
                // Удалить можно только активную И неоплаченную
                actionButton += `<button class="outline secondary btn-icon" onclick="deleteRental(${rental.id})">Удалить</button>`;
            }

            row.innerHTML = `
                <td>${rental.id}</td>
                <td>${serial}</td>
                <td>${rental.renter_name}<br><small>${rental.renter_phone}</small></td>
                <td><small>${formatDate(rental.started_at)}</small></td>
                <td><small>${formatDate(rental.planned_end_at)}</small></td>
                <td>${formatMoney(rental.total_amount)}</td>
                <td>${paidBadge}</td>
                <td>${statusBadge}</td>
                <td style="white-space: nowrap;"><div class="actions-cell">${actionButton}</div></td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load rentals:', err);
    }
}

async function returnRental(id) {
    if (!confirm('Вернуть велосипед (закрыть аренду #' + id + ')?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch('/rentals/' + id + '/return', {
            method: 'POST',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            alert('Ошибка: ' + (err.error || response.status));
            return;
        }

        loadRentals(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

async function markRentalPaid(id) {
    const token = getToken();

    try {
        const response = await fetch('/rentals/' + id + '/pay', {
            method: 'POST',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            alert('Ошибка: ' + (err.error || response.status));
            return;
        }

        loadRentals(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

function formatDate(isoString) {
    if (!isoString) return '—';
    const d = new Date(isoString);
    return d.toLocaleString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    });
}

// ============ Breakdowns page (owner) ============

async function loadBreakdownsPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    const user = await checkRole(token, 'owner');
    if (!user) return;

    loadUserInfo(token);
    loadBreakdowns(token);
}

async function loadBreakdowns(token) {
    const loading = document.getElementById('breakdowns-loading');
    const errorDiv = document.getElementById('breakdowns-error');
    const table = document.getElementById('breakdowns-table');
    const tbody = document.getElementById('breakdowns-tbody');
    const empty = document.getElementById('breakdowns-empty');
    const unfixedOnly = document.getElementById('show-unfixed-only');

    // Выбираем URL в зависимости от чекбокса
    const url = unfixedOnly && unfixedOnly.checked ? '/breakdowns/unfixed' : '/breakdowns';

    try {
        const [breakdownsResp, bikesResp] = await Promise.all([
            fetch(url, { headers: { 'Authorization': 'Bearer ' + token } }),
            fetch('/bikes', { headers: { 'Authorization': 'Bearer ' + token } }),
        ]);

        loading.style.display = 'none';

        if (breakdownsResp.status === 401 || bikesResp.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!breakdownsResp.ok || !bikesResp.ok) {
            throw new Error('Ошибка загрузки данных');
        }

        const breakdowns = await breakdownsResp.json();
        const bikes = await bikesResp.json();

        // Map bike_id → serial_number
        const serialByBikeId = {};
        for (const bike of bikes) {
            serialByBikeId[bike.id] = bike.serial_number;
        }

        if (breakdowns.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

        table.style.display = 'table';
        empty.style.display = 'none';
        tbody.innerHTML = '';

        for (const b of breakdowns) {
            const row = document.createElement('tr');

            const serial = serialByBikeId[b.bike_id] || '#' + b.bike_id;

            const statusBadge = b.is_fixed
                ? '<span style="color: green;">Починен</span>'
                : '<span style="color: red;">Не починен</span>';

            let actionButton = '';
            if (!b.is_fixed) {
                actionButton = `<button class="outline btn-sm" onclick="markBreakdownFixed(${b.id})">Починен</button> `;
            }
            actionButton += `<button class="outline btn-sm" onclick="openEditBreakdownForm(${b.id})">Редактировать</button>`;
            if (!b.is_fixed) {
                actionButton += ` <button class="outline secondary btn-icon" onclick="deleteBreakdown(${b.id})">Удалить</button>`;
            }

            row.innerHTML = `
                <td>${b.id}</td>
                <td>${serial}</td>
                <td>${escapeHtml(b.reason)}</td>
                <td>${formatMoney(b.cost)}</td>
                <td><small>${formatDate(b.broken_at)}</small></td>
                <td>${statusBadge}</td>
                <td style="white-space: nowrap;"><div class="actions-cell">${actionButton}</div></td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load breakdowns:', err);
    }
}

async function markBreakdownFixed(id) {
    if (!confirm('Отметить поломку #' + id + ' как починенную?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch('/breakdowns/' + id + '/fix', {
            method: 'POST',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok && response.status !== 204) {
            const err = await response.json();
            alert('Ошибка: ' + (err.error || response.status));
            return;
        }

        loadBreakdowns(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

// ============ Helper: escapeHtml ============

function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

// ============ Payouts page (owner) ============

async function loadPayoutsPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    const user = await checkRole(token, 'owner');
    if (!user) return;

    loadUserInfo(token);

    // Устанавливаем текущий месяц в input по умолчанию
    const monthInput = document.getElementById('preview-month');
    if (monthInput && !monthInput.value) {
        const now = new Date();
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0');
        monthInput.value = `${year}-${month}`;
    }

    loadPayouts(token);
}

async function loadPreview() {
    const month = document.getElementById('preview-month').value;
    if (!month) {
        alert('Выберите месяц');
        return;
    }

    const token = getToken();
    const loading = document.getElementById('preview-loading');
    const errorDiv = document.getElementById('preview-error');
    const resultDiv = document.getElementById('preview-result');

    loading.style.display = 'block';
    errorDiv.style.display = 'none';
    resultDiv.style.display = 'none';

    try {
        const response = await fetch('/payouts/preview?month=' + encodeURIComponent(month), {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            errorDiv.textContent = err.error || 'Ошибка загрузки preview';
            errorDiv.style.display = 'block';
            return;
        }

        const data = await response.json();

        document.getElementById('preview-revenue').textContent = formatMoney(data.total_revenue);
        document.getElementById('preview-share').textContent = formatMoney(data.partner_share);
        resultDiv.style.display = 'block';

        // Сохраняем месяц для фиксации
        resultDiv.dataset.month = month;
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка сети: ' + err.message;
        errorDiv.style.display = 'block';
    }
}

async function fixPayout() {
    const resultDiv = document.getElementById('preview-result');
    const month = resultDiv.dataset.month;

    if (!month) {
        alert('Сначала посчитайте preview');
        return;
    }

    if (!confirm('Зафиксировать выплату за ' + month + '? После этого изменить нельзя.')) {
        return;
    }

    const token = getToken();
    const errorDiv = document.getElementById('preview-fix-error');
    errorDiv.style.display = 'none';

    try {
        const response = await fetch('/payouts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + token,
            },
            body: JSON.stringify({ month: month }),
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (response.status === 409) {
            errorDiv.textContent = 'Выплата за этот месяц уже существует';
            errorDiv.style.display = 'block';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            errorDiv.textContent = err.error || 'Ошибка';
            errorDiv.style.display = 'block';
            return;
        }

        // Скрываем preview, обновляем список
        resultDiv.style.display = 'none';
        loadPayouts(token);
    } catch (err) {
        errorDiv.textContent = 'Ошибка сети: ' + err.message;
        errorDiv.style.display = 'block';
    }
}

async function loadPayouts(token) {
    const loading = document.getElementById('payouts-loading');
    const errorDiv = document.getElementById('payouts-error');
    const table = document.getElementById('payouts-table');
    const tbody = document.getElementById('payouts-tbody');
    const empty = document.getElementById('payouts-empty');

    try {
        const response = await fetch('/payouts', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            throw new Error('Ошибка загрузки: ' + response.status);
        }

        const payouts = await response.json();

        if (payouts.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

        table.style.display = 'table';
        empty.style.display = 'none';
        tbody.innerHTML = '';

        for (const p of payouts) {
            const row = document.createElement('tr');

            const period = formatPeriod(p.period_start, p.period_end);
            const statusBadge = p.is_paid
                ? '<span style="color: green;">Выплачено</span>'
                : '<span style="color: orange;">Ожидает</span>';

            let actionButton = '';
            if (!p.is_paid) {
                actionButton = `<button class="outline btn-sm" onclick="markPayoutPaid(${p.id})">Выплачено</button> `;
                actionButton += `<button class="outline secondary btn-icon" onclick="deletePayout(${p.id})">Удалить</button>`;
            }

            row.innerHTML = `
                <td>${p.id}</td>
                <td>${period}</td>
                <td>${formatMoney(p.total_revenue)}</td>
                <td><strong style="color: green;">${formatMoney(p.partner_share)}</strong></td>
                <td>${statusBadge}</td>
                <td style="white-space: nowrap;"><div class="actions-cell">${actionButton}</div></td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load payouts:', err);
    }
}

async function markPayoutPaid(id) {
    if (!confirm('Отметить выплату #' + id + ' как выполненную?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch('/payouts/' + id + '/pay', {
            method: 'POST',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            alert('Ошибка: ' + (err.error || response.status));
            return;
        }

        loadPayouts(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

function formatPeriod(startISO, endISO) {
    // period_start = 2026-10-01, period_end = 2026-11-01
    // Показываем как "октябрь 2026"
    const start = new Date(startISO);
    const months = [
        'январь', 'февраль', 'март', 'апрель', 'май', 'июнь',
        'июль', 'август', 'сентябрь', 'октябрь', 'ноябрь', 'декабрь'
    ];
    return months[start.getMonth()] + ' ' + start.getFullYear();
}

// ============ Partner Bikes page ============

async function loadPartnerBikesPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    const user = await checkRole(token, 'partner');
    if (!user) return;

    loadUserInfo(token);
    loadPartnerBikes(token);
}

async function loadPartnerBikes(token) {
    const loading = document.getElementById('bikes-loading');
    const errorDiv = document.getElementById('bikes-error');
    const table = document.getElementById('bikes-table');
    const tbody = document.getElementById('bikes-tbody');
    const empty = document.getElementById('bikes-empty');

    try {
        const response = await fetch('/partner/bikes', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            throw new Error('Ошибка загрузки: ' + response.status);
        }

        const bikes = await response.json();

        if (bikes.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

        empty.style.display = 'none';
        table.style.display = 'table';
        tbody.innerHTML = '';

        for (const bike of bikes) {
            const row = document.createElement('tr');

            let statusBadge = '';
            if (bike.is_broken) {
                statusBadge = '<span style="color: red;">Сломан</span>';
            } else if (bike.is_rented) {
                statusBadge = '<span style="color: orange;">Сдан</span>';
            } else {
                statusBadge = '<span style="color: green;">Свободен</span>';
            }

            // Формируем фото-ячейку
            let photoCell = '<span style="color: gray;">—</span>';
            if (bike.photo_url) {
                photoCell = `<img src="${bike.photo_url}" alt="Фото" style="width: 50px; height: 50px; object-fit: cover; border-radius: 4px;">`;
            }

            row.innerHTML = `
                <td>${photoCell}</td>
                <td>${escapeHtml(bike.serial_number)}</td>
                <td>${escapeHtml(bike.model || '—')}</td>
                <td>${statusBadge}</td>
                <td style="white-space: nowrap;">
                    <div class="actions-cell">
                        <button class="outline btn-sm" onclick="showBikeDetails(${bike.id}, '${escapeHtml(bike.serial_number)}')">
                            Подробнее
                        </button>
                    </div>
                </td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load partner bikes:', err);
    }
}

async function showBikeDetails(bikeId, serial) {
    const details = document.getElementById('bike-details');
    const title = document.getElementById('bike-details-title');

    title.textContent = `Велосипед ${serial}`;
    details.style.display = 'block';
    details.scrollIntoView({ behavior: 'smooth' });

    const token = getToken();
    const role = getUserRole();
    const baseUrl = role === 'partner' ? '/partner' : '';

    // Показываем заглушки на время загрузки
    document.getElementById('bike-details-serial').textContent = serial;
    document.getElementById('bike-details-model').textContent = 'Загрузка...';
    document.getElementById('bike-details-status').textContent = '—';
    document.getElementById('bike-details-created').textContent = '—';
    document.getElementById('bike-details-photo-container').innerHTML =
        '<div style="display: flex; align-items: center; justify-content: center; ' +
        'color: var(--pico-muted-color);">Загрузка...</div>';

    // Загружаем данные велосипеда
    try {
        const bikeResp = await fetch(`${baseUrl}/bikes/${bikeId}`, {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (!bikeResp.ok) {
            throw new Error('Не удалось загрузить велосипед');
        }

        const bike = await bikeResp.json();

        // Заполняем информацию
        document.getElementById('bike-details-serial').textContent = bike.serial_number;
        document.getElementById('bike-details-model').textContent = bike.model || '—';
        document.getElementById('bike-details-created').textContent = formatDate(bike.created_at);

        // Статус
        let statusHtml = '';
        if (bike.is_broken) {
            statusHtml = '<span style="color: #c0392b; font-weight: 600;">Сломан</span>';
        } else if (bike.is_rented) {
            statusHtml = '<span style="color: #e67e22; font-weight: 600;">Сдан</span>';
        } else {
            statusHtml = '<span style="color: #27ae60; font-weight: 600;">Свободен</span>';
        }
        document.getElementById('bike-details-status').innerHTML = statusHtml;

        // Фото
        const photoContainer = document.getElementById('bike-details-photo-container');
        if (bike.photo_url) {
            photoContainer.innerHTML = `
                <a href="${bike.photo_url}" target="_blank" title="Открыть полное фото">
                    <img src="${bike.photo_url}" alt="${escapeHtml(bike.serial_number)}" style="cursor: zoom-in;">
                </a>
            `;
        } else {
            photoContainer.innerHTML = `
                <div style="display: flex; align-items: center; justify-content: center; color: var(--pico-muted-color); background: var(--pico-muted-border-color);">
                    Нет фото
                </div>
            `;
        }
    } catch (err) {
        console.error('Failed to load bike details:', err);
        document.getElementById('bike-details-model').textContent = '—';
        document.getElementById('bike-details-photo-container').innerHTML =
            '<div style="width: 240px; height: 240px; background: var(--pico-muted-border-color); border-radius: 8px; display: flex; align-items: center; justify-content: center; color: var(--pico-muted-color);">Ошибка загрузки</div>';
    }

    // Загружаем поломки (как было)
    const loading = document.getElementById('breakdowns-loading');
    const errorDiv = document.getElementById('breakdowns-error');
    const table = document.getElementById('breakdowns-table');
    const tbody = document.getElementById('breakdowns-tbody');
    const empty = document.getElementById('breakdowns-empty');

    loading.style.display = 'block';
    errorDiv.style.display = 'none';
    table.style.display = 'none';
    empty.style.display = 'none';

    // Эндпоинт поломок — зависит от роли
    const breakdownsUrl = role === 'partner'
        ? `/partner/bikes/${bikeId}/breakdowns`
        : `/bikes/${bikeId}/breakdowns`;

    try {
        const response = await fetch(breakdownsUrl, {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (!response.ok) {
            throw new Error('Ошибка загрузки: ' + response.status);
        }

        const breakdowns = await response.json();

        if (breakdowns.length === 0) {
            empty.style.display = 'block';
            return;
        }

        table.style.display = 'table';
        tbody.innerHTML = '';

        for (const b of breakdowns) {
            const row = document.createElement('tr');

            const statusBadge = b.is_fixed
                ? '<span style="color: #27ae60; font-weight: 600;">Починен</span>'
                : '<span style="color: #c0392b; font-weight: 600;">Не починен</span>';

            row.innerHTML = `
                <td>${escapeHtml(b.reason)}</td>
                <td>${formatMoney(b.cost)}</td>
                <td><small>${formatDate(b.broken_at)}</small></td>
                <td>${statusBadge}</td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load breakdowns:', err);
    }
}

function closeBikeDetails() {
    document.getElementById('bike-details').style.display = 'none';
}

// ============ Partner Rentals page ============

async function loadPartnerRentalsPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    const user = await checkRole(token, 'partner');
    if (!user) return;

    loadUserInfo(token);
    loadPartnerRentals(token);
}

async function loadPartnerRentals(token) {
    const loading = document.getElementById('rentals-loading');
    const errorDiv = document.getElementById('rentals-error');
    const table = document.getElementById('rentals-table');
    const tbody = document.getElementById('rentals-tbody');
    const empty = document.getElementById('rentals-empty');

    try {
        const response = await fetch('/partner/rentals', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            throw new Error('Ошибка загрузки: ' + response.status);
        }

        const rentals = await response.json();

        if (rentals.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

        empty.style.display = 'none';
        table.style.display = 'table';
        tbody.innerHTML = '';

        for (const rental of rentals) {
            const row = document.createElement('tr');

            let statusBadge;
            if (rental.returned_at) {
                statusBadge = '<span style="color: green;">Возвращён</span>';
            } else {
                const plannedEnd = new Date(rental.planned_end_at);
                if (plannedEnd < new Date()) {
                    statusBadge = '<span style="color: red;">Просрочен</span>';
                } else {
                    statusBadge = '<span style="color: orange;">Активен</span>';
                }
            }

            const paidBadge = rental.is_paid
                ? '<span style="color: green;">Оплачено</span>'
                : '<span style="color: red;">Не оплачено</span>';

            row.innerHTML = `
                <td>${escapeHtml(rental.bike_serial)}</td>
                <td>${escapeHtml(rental.renter_name)}<br><small>${escapeHtml(rental.renter_phone)}</small></td>
                <td><small>${formatDate(rental.started_at)}</small></td>
                <td><small>${formatDate(rental.planned_end_at)}</small></td>
                <td>${formatMoney(rental.total_amount)}</td>
                <td>${paidBadge}</td>
                <td>${statusBadge}</td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load partner rentals:', err);
    }
}

// ============ Partner Breakdowns page ============

async function loadPartnerBreakdownsPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    const user = await checkRole(token, 'partner');
    if (!user) return;

    loadUserInfo(token);
    loadPartnerBreakdowns(token);
}

async function loadPartnerBreakdowns(token) {
    const loading = document.getElementById('breakdowns-loading');
    const errorDiv = document.getElementById('breakdowns-error');
    const table = document.getElementById('breakdowns-table');
    const tbody = document.getElementById('breakdowns-tbody');
    const empty = document.getElementById('breakdowns-empty');

    try {
        // Загружаем breakdowns + bikes (для serial)
        const [breakdownsResp, bikesResp] = await Promise.all([
            fetch('/partner/breakdowns', { headers: { 'Authorization': 'Bearer ' + token } }),
            fetch('/partner/bikes', { headers: { 'Authorization': 'Bearer ' + token } }),
        ]);

        loading.style.display = 'none';

        if (!breakdownsResp.ok || !bikesResp.ok) {
            throw new Error('Ошибка загрузки данных');
        }

        const breakdowns = await breakdownsResp.json();
        const bikes = await bikesResp.json();

        const serialByBikeId = {};
        for (const bike of bikes) {
            serialByBikeId[bike.id] = bike.serial_number;
        }

        if (breakdowns.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

        table.style.display = 'table';
        empty.style.display = 'none';
        tbody.innerHTML = '';

        for (const b of breakdowns) {
            const row = document.createElement('tr');

            const serial = serialByBikeId[b.bike_id] || '#' + b.bike_id;
            const statusBadge = b.is_fixed
                ? '<span style="color: green;">Починен</span>'
                : '<span style="color: red;">Не починен</span>';

            row.innerHTML = `
                <td>${escapeHtml(serial)}</td>
                <td>${escapeHtml(b.reason)}</td>
                <td>${formatMoney(b.cost)}</td>
                <td><small>${formatDate(b.broken_at)}</small></td>
                <td>${statusBadge}</td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load partner breakdowns:', err);
    }
}

// ============ Partner Payouts page ============

async function loadPartnerPayoutsPage() {
    const token = getToken();
    if (!token) {
        window.location.href = '/web/login';
        return;
    }

    const user = await checkRole(token, 'partner');
    if (!user) return;

    loadUserInfo(token);

    // Текущий месяц по умолчанию
    const monthInput = document.getElementById('preview-month');
    if (monthInput && !monthInput.value) {
        const now = new Date();
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0');
        monthInput.value = `${year}-${month}`;
    }

    loadPartnerPayouts(token);
}

async function loadPartnerPreview() {
    const month = document.getElementById('preview-month').value;
    if (!month) {
        alert('Выберите месяц');
        return;
    }

    const token = getToken();
    const loading = document.getElementById('preview-loading');
    const errorDiv = document.getElementById('preview-error');
    const resultDiv = document.getElementById('preview-result');

    loading.style.display = 'block';
    errorDiv.style.display = 'none';
    resultDiv.style.display = 'none';

    try {
        const response = await fetch('/partner/payouts/preview?month=' + encodeURIComponent(month), {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (!response.ok) {
            const err = await response.json();
            errorDiv.textContent = err.error || 'Ошибка';
            errorDiv.style.display = 'block';
            return;
        }

        const data = await response.json();
        document.getElementById('preview-revenue').textContent = formatMoney(data.total_revenue);
        document.getElementById('preview-share').textContent = formatMoney(data.partner_share);
        resultDiv.style.display = 'block';
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка сети: ' + err.message;
        errorDiv.style.display = 'block';
    }
}

async function loadPartnerPayouts(token) {
    const loading = document.getElementById('payouts-loading');
    const errorDiv = document.getElementById('payouts-error');
    const table = document.getElementById('payouts-table');
    const tbody = document.getElementById('payouts-tbody');
    const empty = document.getElementById('payouts-empty');

    try {
        const response = await fetch('/partner/payouts', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        loading.style.display = 'none';

        if (!response.ok) {
            throw new Error('Ошибка загрузки: ' + response.status);
        }

        const payouts = await response.json();

        if (payouts.length === 0) {
            empty.style.display = 'block';
            table.style.display = 'none';
            return;
        }

        table.style.display = 'table';
        empty.style.display = 'none';
        tbody.innerHTML = '';

        for (const p of payouts) {
            const row = document.createElement('tr');

            const period = formatPeriod(p.period_start, p.period_end);
            const statusBadge = p.is_paid
                ? '<span style="color: green;">Выплачено</span>'
                : '<span style="color: orange;">Ожидает</span>';

            row.innerHTML = `
                <td>${period}</td>
                <td>${formatMoney(p.total_revenue)}</td>
                <td><strong style="color: green;">${formatMoney(p.partner_share)}</strong></td>
                <td>${statusBadge}</td>
            `;
            tbody.appendChild(row);
        }
    } catch (err) {
        loading.style.display = 'none';
        errorDiv.textContent = 'Ошибка: ' + err.message;
        errorDiv.style.display = 'block';
        console.error('Failed to load partner payouts:', err);
    }
}

// ============ Bikes: form (create / edit) ============

function openCreateBikeForm() {
    // Сброс формы
    document.getElementById('bike-form').reset();
    document.getElementById('bike-form-id').value = '';
    document.getElementById('bike-form-title').textContent = 'Новый велосипед';
    document.getElementById('bike-form-submit').textContent = 'Сохранить';
    document.getElementById('bike-form-error').style.display = 'none';

    // Сброс фото
    document.getElementById('bike-form-photo').value = '';
    document.getElementById('bike-photo-preview').innerHTML = '';
    document.getElementById('bike-photo-preview').style.display = 'none';

    document.getElementById('bike-form-container').style.display = 'block';


}

async function openEditBikeForm(id) {
    const token = getToken();
    const errorDiv = document.getElementById('bike-form-error');
    errorDiv.style.display = 'none';

    try {
        const response = await fetch('/bikes/' + id, {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (!response.ok) {
            alert('Не удалось загрузить велосипед');
            return;
        }

        const bike = await response.json();

        // Заполняем форму
        document.getElementById('bike-form-id').value = bike.id;
        document.getElementById('bike-form-serial').value = bike.serial_number;
        document.getElementById('bike-form-model').value = bike.model || '';
        document.getElementById('bike-form-comment').value = bike.comment || '';

        // Показываем текущее фото (если есть)
        const preview = document.getElementById('bike-photo-preview');
        preview.innerHTML = '';
        preview.style.display = 'block';

        if (bike.photo_url) {
            preview.innerHTML = `
                <div style="display: flex; align-items: center; gap: 1rem; flex-wrap: wrap;">
                    <a href="${bike.photo_url}" target="_blank" title="Открыть полное фото">
                        <img src="${bike.photo_url}" alt="${escapeHtml(bike.serial_number)}" 
                             style="width: 120px; height: 120px; object-fit: cover; border-radius: 6px; border: 1px solid var(--pico-muted-border-color); cursor: zoom-in;">
                    </a>
                    <button type="button" class="secondary outline btn-sm" onclick="removeBikePhoto(${bike.id})">
                        Удалить фото
                    </button>
                </div>
            `;
        } else {
            preview.innerHTML = '<small style="color: var(--pico-muted-color);">Фото не загружено</small>';
        }

        // Меняем заголовок и кнопку
        document.getElementById('bike-form-title').textContent = `Редактировать велосипед ${bike.serial_number}`;
        document.getElementById('bike-form-submit').textContent = 'Сохранить изменения';

        // Показываем
        const container = document.getElementById('bike-form-container');
        container.style.display = 'block';
        container.scrollIntoView({ behavior: 'smooth' });
    } catch (err) {
        alert('Ошибка: ' + err.message);
    }
}

function closeBikeForm() {
    document.getElementById('bike-form-container').style.display = 'none';
    document.getElementById('bike-form').reset();
    document.getElementById('bike-form-id').value = '';
    document.getElementById('bike-form-photo').value = '';
    document.getElementById('bike-photo-preview').innerHTML = '';
    document.getElementById('bike-photo-preview').style.display = 'none';
}

async function handleBikeFormSubmit(event) {
    event.preventDefault();

    const token = getToken();
    const form = event.target;
    const errorDiv = document.getElementById('bike-form-error');
    errorDiv.style.display = 'none';

    const id = document.getElementById('bike-form-id').value;
    const isEdit = id !== '';

    const body = {
        serial_number: form.serial_number.value.trim(),
        model: form.model.value.trim() || null,
        comment: form.comment.value.trim() || null,
    };

    const url = isEdit ? `/bikes/${id}` : '/bikes';
    const method = isEdit ? 'PUT' : 'POST';

    try {
        // 1. Сохраняем велосипед (создаём или обновляем)
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + token,
            },
            body: JSON.stringify(body),
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            errorDiv.textContent = err.error || 'Ошибка сохранения';
            errorDiv.style.display = 'block';
            return;
        }

        const savedBike = await response.json();
        const bikeId = savedBike.id;

        // 2. Если выбран файл — загружаем фото
        const photoInput = document.getElementById('bike-form-photo');
        if (photoInput.files.length > 0) {
            const file = photoInput.files[0];

            // Проверка размера (клиентская)
            if (file.size > 5 * 1024 * 1024) {
                errorDiv.textContent = 'Фото слишком большое (макс 5 MB)';
                errorDiv.style.display = 'block';
                return;
            }

            const formData = new FormData();
            formData.append('photo', file);

            const photoResponse = await fetch(`/bikes/${bikeId}/photo`, {
                method: 'POST',
                headers: {
                    'Authorization': 'Bearer ' + token,
                    // Content-Type НЕ ставим — браузер сам поставит multipart с boundary
                },
                body: formData,
            });

            if (!photoResponse.ok) {
                const err = await photoResponse.json();
                errorDiv.textContent = 'Велосипед сохранён, но фото не загружено: ' + (err.error || photoResponse.status);
                errorDiv.style.display = 'block';
                // Не падаем — велосипед уже сохранён
            }
        }

        closeBikeForm();
        loadBikes(token, true);
    } catch (err) {
        errorDiv.textContent = 'Ошибка сети: ' + err.message;
        errorDiv.style.display = 'block';
    }
}

// ============ Breakdowns: form (create / edit) ============

async function openCreateBreakdownForm() {
    // Сброс формы
    document.getElementById('breakdown-form').reset();
    document.getElementById('breakdown-form-id').value = '';
    document.getElementById('breakdown-form-title').textContent = 'Новая поломка';
    document.getElementById('breakdown-form-submit').textContent = 'Сохранить';
    document.getElementById('breakdown-form-error').style.display = 'none';
    document.getElementById('breakdown-form-cost').value = '0';

    // Загружаем ВСЕ велосипеды (не только свободные)
    await fillBikesDropdown(document.getElementById('breakdown-form-bike'));

    document.getElementById('breakdown-form-container').style.display = 'block';
}

async function openEditBreakdownForm(id) {
    const token = getToken();
    const errorDiv = document.getElementById('breakdown-form-error');
    errorDiv.style.display = 'none';

    try {
        const response = await fetch('/breakdowns/' + id, {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (!response.ok) {
            alert('Не удалось загрузить поломку');
            return;
        }

        const b = await response.json();

        // Заполняем форму
        document.getElementById('breakdown-form-id').value = b.id;

        // Заполняем список велосипедов (все)
        await fillBikesDropdown(document.getElementById('breakdown-form-bike'));
        document.getElementById('breakdown-form-bike').value = b.bike_id;

        document.getElementById('breakdown-form-reason').value = b.reason;
        document.getElementById('breakdown-form-cost').value = b.cost;
        document.getElementById('breakdown-form-comment').value = b.comment || '';

        // Заголовок и кнопка
        document.getElementById('breakdown-form-title').textContent = 'Редактировать поломку';
        document.getElementById('breakdown-form-submit').textContent = 'Сохранить изменения';

        // Показываем
        const container = document.getElementById('breakdown-form-container');
        container.style.display = 'block';
        container.scrollIntoView({ behavior: 'smooth' });
    } catch (err) {
        alert('Ошибка: ' + err.message);
    }
}

function closeBreakdownForm() {
    document.getElementById('breakdown-form-container').style.display = 'none';
    document.getElementById('breakdown-form').reset();
    document.getElementById('breakdown-form-id').value = '';
}

async function handleBreakdownFormSubmit(event) {
    event.preventDefault();

    const token = getToken();
    const form = event.target;
    const errorDiv = document.getElementById('breakdown-form-error');
    errorDiv.style.display = 'none';

    const id = document.getElementById('breakdown-form-id').value;
    const isEdit = id !== '';

    const body = {
        bike_id: parseInt(form.bike_id.value, 10),
        reason: form.reason.value.trim(),
        cost: parseInt(form.cost.value || '0', 10),
        comment: form.comment.value.trim() || null,
    };

    const url = isEdit ? `/breakdowns/${id}` : '/breakdowns';
    const method = isEdit ? 'PUT' : 'POST';

    try {
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + token,
            },
            body: JSON.stringify(body),
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            errorDiv.textContent = err.error || 'Ошибка сохранения';
            errorDiv.style.display = 'block';
            return;
        }

        closeBreakdownForm();
        loadBreakdowns(token);
    } catch (err) {
        errorDiv.textContent = 'Ошибка сети: ' + err.message;
        errorDiv.style.display = 'block';
    }
}

// Хелпер: заполняет <select> всеми велосипедами
async function fillBikesDropdown(select) {
    const token = getToken();

    select.innerHTML = '<option value="">— Выберите велосипед —</option>';

    try {
        const response = await fetch('/bikes', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (!response.ok) return;

        const bikes = await response.json();

        if (bikes.length === 0) {
            select.innerHTML = '<option value="">— Нет велосипедов —</option>';
            return;
        }

        for (const bike of bikes) {
            const option = document.createElement('option');
            option.value = bike.id;
            let label = bike.serial_number;
            if (bike.model) label += ' — ' + bike.model;
            if (bike.is_broken) label += ' (сломан)';
            option.textContent = label;
            select.appendChild(option);
        }
    } catch (err) {
        console.error('Failed to load bikes:', err);
    }
}

// ============ Rentals: form (create / edit) ============

async function openCreateRentalForm() {
    document.getElementById('rental-form').reset();
    document.getElementById('rental-form-id').value = '';
    document.getElementById('rental-form-title').textContent = 'Новая аренда';
    document.getElementById('rental-form-submit').textContent = 'Создать аренду';
    document.getElementById('rental-form-error').style.display = 'none';

    // Разблокируем все поля для создания
    document.getElementById('rental-form-started').disabled = false;
    document.getElementById('rental-form-prepayment').disabled = false;
    document.getElementById('rental-form-deposit').disabled = false;
    document.getElementById('rental-form-bike').disabled = false;

    // Только свободные велосипеды
    await fillAvailableBikesDropdown(document.getElementById('rental-form-bike'));

    // Настраиваем превью серийника
    document.getElementById('rental-form-bike').onchange = function() {
        const selected = this.options[this.selectedIndex];
        document.getElementById('rental-form-serial-preview').value = selected.dataset.serial || '';
    };

    document.getElementById('rental-form-container').style.display = 'block';
}

async function openEditRentalForm(id) {
    const token = getToken();
    const errorDiv = document.getElementById('rental-form-error');
    errorDiv.style.display = 'none';

    try {
        const response = await fetch('/rentals/' + id, {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (!response.ok) {
            alert('Не удалось загрузить аренду');
            return;
        }

        const r = await response.json();

        // Проверяем: аренда активна?
        if (r.returned_at) {
            alert('Нельзя редактировать возвращённую аренду');
            return;
        }

        // Заполняем форму
        document.getElementById('rental-form-id').value = r.id;

        // Загружаем ВСЕ велосипеды + выбираем текущий
        await fillAllBikesDropdown(document.getElementById('rental-form-bike'));
        document.getElementById('rental-form-bike').value = r.bike_id;

        // Serial preview
        const selected = document.getElementById('rental-form-bike').options[document.getElementById('rental-form-bike').selectedIndex];
        document.getElementById('rental-form-serial-preview').value = selected.dataset.serial || '';

        document.getElementById('rental-form-name').value = r.renter_name;
        document.getElementById('rental-form-phone').value = r.renter_phone;
        document.getElementById('rental-form-started').value = toDatetimeLocal(r.started_at);
        document.getElementById('rental-form-end').value = toDatetimeLocal(r.planned_end_at);
        document.getElementById('rental-form-total').value = r.total_amount;
        document.getElementById('rental-form-prepayment').value = r.prepayment || '';
        document.getElementById('rental-form-deposit').value = r.deposit || '';
        document.getElementById('rental-form-comment').value = r.comment || '';

        // Блокируем неизменяемые поля
        document.getElementById('rental-form-started').disabled = true;
        document.getElementById('rental-form-prepayment').disabled = true;
        document.getElementById('rental-form-deposit').disabled = true;
        document.getElementById('rental-form-bike').disabled = true;

        // Заголовок и кнопка
        document.getElementById('rental-form-title').textContent = 'Редактировать аренду';
        document.getElementById('rental-form-submit').textContent = 'Сохранить изменения';

        const container = document.getElementById('rental-form-container');
        container.style.display = 'block';
        container.scrollIntoView({ behavior: 'smooth' });
    } catch (err) {
        alert('Ошибка: ' + err.message);
    }
}

function closeRentalForm() {
    document.getElementById('rental-form-container').style.display = 'none';
    document.getElementById('rental-form').reset();
    document.getElementById('rental-form-id').value = '';
}

async function handleRentalFormSubmit(event) {
    event.preventDefault();

    const token = getToken();
    const form = event.target;
    const errorDiv = document.getElementById('rental-form-error');
    errorDiv.style.display = 'none';

    const id = document.getElementById('rental-form-id').value;
    const isEdit = id !== '';

    // Для создания
    if (!isEdit) {
        const body = {
            bike_id: parseInt(form.bike_id.value, 10),
            renter_name: form.renter_name.value.trim(),
            renter_phone: form.renter_phone.value.trim(),
            planned_end_at: new Date(form.planned_end_at.value).toISOString(),
            total_amount: parseInt(form.total_amount.value, 10),
        };

        if (form.started_at.value) {
            body.started_at = new Date(form.started_at.value).toISOString();
        }
        if (form.prepayment.value) {
            body.prepayment = parseInt(form.prepayment.value, 10);
        }
        if (form.deposit.value) {
            body.deposit = parseInt(form.deposit.value, 10);
        }
        if (form.comment.value.trim()) {
            body.comment = form.comment.value.trim();
        }

        return sendRentalRequest('/rentals', 'POST', body, token, errorDiv);
    }

    // Для редактирования — только изменяемые поля
    const body = {
        renter_name: form.renter_name.value.trim(),
        renter_phone: form.renter_phone.value.trim(),
        planned_end_at: new Date(form.planned_end_at.value).toISOString(),
        total_amount: parseInt(form.total_amount.value, 10),
        comment: form.comment.value.trim() || null,
    };

    return sendRentalRequest(`/rentals/${id}`, 'PUT', body, token, errorDiv);
}

async function sendRentalRequest(url, method, body, token, errorDiv) {
    try {
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + token,
            },
            body: JSON.stringify(body),
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok) {
            const err = await response.json();
            errorDiv.textContent = err.error || 'Ошибка сохранения';
            errorDiv.style.display = 'block';
            return;
        }

        closeRentalForm();
        loadRentals(token);
    } catch (err) {
        errorDiv.textContent = 'Ошибка сети: ' + err.message;
        errorDiv.style.display = 'block';
    }
}

// Хелпер: только свободные велосипеды (для создания)
async function fillAvailableBikesDropdown(select) {
    const token = getToken();
    select.innerHTML = '<option value="">— Выберите велосипед —</option>';

    try {
        const response = await fetch('/bikes', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (!response.ok) return;

        const bikes = await response.json();
        const available = bikes.filter(b => !b.is_rented && !b.is_broken);

        if (available.length === 0) {
            select.innerHTML = '<option value="">— Нет свободных велосипедов —</option>';
            return;
        }

        for (const bike of available) {
            const option = document.createElement('option');
            option.value = bike.id;
            option.textContent = `${bike.serial_number} — ${bike.model || 'без модели'}`;
            option.dataset.serial = bike.serial_number;
            select.appendChild(option);
        }
    } catch (err) {
        console.error('Failed to load bikes:', err);
    }
}

// Хелпер: все велосипеды (для редактирования)
async function fillAllBikesDropdown(select) {
    const token = getToken();
    select.innerHTML = '<option value="">— Выберите велосипед —</option>';

    try {
        const response = await fetch('/bikes', {
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (!response.ok) return;

        const bikes = await response.json();

        for (const bike of bikes) {
            const option = document.createElement('option');
            option.value = bike.id;
            let label = bike.serial_number;
            if (bike.model) label += ' — ' + bike.model;
            if (bike.is_rented) label += ' (сдан)';
            if (bike.is_broken) label += ' (сломан)';
            option.textContent = label;
            option.dataset.serial = bike.serial_number;
            select.appendChild(option);
        }
    } catch (err) {
        console.error('Failed to load bikes:', err);
    }
}

// Хелпер: ISO → формат для input[type="datetime-local"]
function toDatetimeLocal(isoString) {
    if (!isoString) return '';
    const d = new Date(isoString);
    const pad = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

async function deleteRental(id) {
    if (!confirm('Удалить аренду #' + id + '?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch('/rentals/' + id, {
            method: 'DELETE',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (response.status === 409) {
            const err = await response.json();
            alert(err.error || 'Нельзя удалить эту аренду');
            return;
        }

        if (!response.ok && response.status !== 204) {
            const err = await response.json();
            alert('Ошибка удаления: ' + (err.error || response.status));
            return;
        }

        loadRentals(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

async function deleteBreakdown(id) {
    if (!confirm('Удалить поломку #' + id + '?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch('/breakdowns/' + id, {
            method: 'DELETE',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (response.status === 409) {
            const err = await response.json();
            alert(err.error || 'Нельзя удалить эту поломку');
            return;
        }

        if (!response.ok && response.status !== 204) {
            const err = await response.json();
            alert('Ошибка удаления: ' + (err.error || response.status));
            return;
        }

        loadBreakdowns(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

async function deletePayout(id) {
    if (!confirm('Удалить выплату #' + id + '?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch('/payouts/' + id, {
            method: 'DELETE',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (response.status === 409) {
            const err = await response.json();
            alert(err.error || 'Нельзя удалить эту выплату');
            return;
        }

        if (!response.ok && response.status !== 204) {
            const err = await response.json();
            alert('Ошибка удаления: ' + (err.error || response.status));
            return;
        }

        loadPayouts(token);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

async function removeBikePhoto(id) {
    if (!confirm('Удалить фото велосипеда?')) {
        return;
    }

    const token = getToken();

    try {
        const response = await fetch(`/bikes/${id}/photo`, {
            method: 'DELETE',
            headers: { 'Authorization': 'Bearer ' + token },
        });

        if (response.status === 401) {
            clearToken();
            window.location.href = '/web/login';
            return;
        }

        if (!response.ok && response.status !== 204) {
            const err = await response.json();
            alert('Ошибка удаления фото: ' + (err.error || response.status));
            return;
        }

        // Обновляем превью
        const preview = document.getElementById('bike-photo-preview');
        preview.innerHTML = '<small style="color: var(--pico-muted-color);">Фото удалено</small>';

        // Обновляем список
        loadBikes(token, true);
    } catch (err) {
        alert('Ошибка сети: ' + err.message);
    }
}

// ============ Mobile nav toggle ============

function toggleNav() {
    const menu = document.getElementById('nav-menu');
    if (!menu) return;
    menu.classList.toggle('open');
}

// Закрываем меню при клике на ссылку (на мобильном)
document.addEventListener('click', (e) => {
    const menu = document.getElementById('nav-menu');
    if (!menu || !menu.classList.contains('open')) return;
    if (e.target.tagName === 'A') {
        menu.classList.remove('open');
    }
});