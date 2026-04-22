package seed

import (
	"database/sql"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
)

func hashPwd(p string) string {
	h := sha256.New()
	h.Write([]byte(p))
	return hex.EncodeToString(h.Sum(nil))
}

func Run(db *sql.DB) error {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		return nil
	}

	log.Println("Seeding database...")

	pwd := hashPwd("user123")
	adminPwd := hashPwd("admin123")

	users := []struct {
		email, nickname, password, name, city, bio, role, interests, userRole2 string
	}{
		{"admin@startup.by", "admin", adminPwd, "Администратор", "Минск", "Главный администратор платформы", "admin", "управление,модерация", "Администратор"},
		{"ivan@startup.by", "ivan_dev", pwd, "Иван Петров", "Минск", "Full-stack разработчик с 5-летним опытом", "user", "программирование,веб,AI", "Разработчик"},
		{"maria@startup.by", "maria_design", pwd, "Мария Козлова", "Гродно", "UI/UX дизайнер, люблю минимализм", "premium", "дизайн,UX,брендинг", "Дизайнер"},
		{"alex@startup.by", "alex_pm", pwd, "Алексей Сидоров", "Брест", "Менеджер проектов в IT", "user", "управление,agile,стартапы", "PM"},
		{"olga@startup.by", "olga_expert", pwd, "Ольга Новикова", "Минск", "Эксперт по венчурным инвестициям", "expert", "финансы,инвестиции,стартапы", "Инвестор"},
		{"dmitry@startup.by", "dmitry_ml", pwd, "Дмитрий Волков", "Витебск", "ML-инженер, работаю с NLP", "user", "AI,ML,данные", "ML-инженер"},
		{"anna@startup.by", "anna_mark", pwd, "Анна Морозова", "Минск", "Маркетолог с опытом в стартапах", "premium", "маркетинг,SMM,аналитика", "Маркетолог"},
		{"sergey@startup.by", "sergey_back", pwd, "Сергей Лебедев", "Гомель", "Backend-разработчик на Go и Python", "user", "Go,Python,микросервисы", "Backend-разработчик"},
		{"elena@startup.by", "elena_data", pwd, "Елена Кузнецова", "Минск", "Аналитик данных", "user", "аналитика,SQL,визуализация", "Аналитик"},
		{"nikita@startup.by", "nikita_front", pwd, "Никита Попов", "Могилёв", "Frontend-разработчик React/Vue", "user", "React,Vue,TypeScript", "Frontend-разработчик"},
		{"kate@startup.by", "kate_hr", pwd, "Екатерина Соколова", "Минск", "HR в IT-компаниях", "user", "HR,рекрутинг,культура", "HR-менеджер"},
		{"pavel@startup.by", "pavel_devops", pwd, "Павел Фёдоров", "Брест", "DevOps-инженер", "user", "DevOps,Docker,AWS", "DevOps"},
		{"yulia@startup.by", "yulia_content", pwd, "Юлия Васильева", "Гродно", "Контент-менеджер", "user", "контент,копирайтинг,SMM", "Контент-менеджер"},
		{"maxim@startup.by", "maxim_sec", pwd, "Максим Зайцев", "Минск", "Специалист по кибербезопасности", "expert", "безопасность,пентест,аудит", "Безопасник"},
		{"natasha@startup.by", "natasha_qa", pwd, "Наталья Белова", "Витебск", "QA-инженер", "user", "тестирование,автоматизация,качество", "QA"},
		{"artem@startup.by", "artem_game", pwd, "Артём Ковалёв", "Минск", "Геймдев-разработчик Unity", "user", "геймдев,Unity,3D", "Геймдев"},
		{"light@startup.by", "light_startup", pwd, "Светлана Романова", "Гомель", "Основатель двух стартапов", "premium", "стартапы,бизнес,продукт", "Основатель"},
		{"viktor@startup.by", "viktor_mobile", pwd, "Виктор Михайлов", "Минск", "Мобильный разработчик iOS/Android", "user", "iOS,Android,Flutter", "Mobile-разработчик"},
		{"daria@startup.by", "daria_ux", pwd, "Дарья Орлова", "Брест", "UX-исследователь", "user", "UX,исследования,интервью", "UX-исследователь"},
		{"roman@startup.by", "roman_cto", pwd, "Роман Николаев", "Минск", "CTO стартапа EduTech", "user", "архитектура,стратегия,лидерство", "CTO"},
		{"polina@startup.by", "polina_ai", pwd, "Полина Андреева", "Витебск", "AI-исследователь в БГУИР", "user", "AI,исследования,публикации", "AI-исследователь"},
	}

	for _, u := range users {
		_, err := db.Exec(
			`INSERT INTO users (email, nickname, password_hash, name, city, bio, role, interests, user_role2) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			u.email, u.nickname, u.password, u.name, u.city, u.bio, u.role, u.interests, u.userRole2,
		)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.nickname, err)
		}
	}

	posts := []struct {
		authorID                                 int
		title, desc, category, tags string
	}{
		{2, "EcoTrack — трекер углеродного следа", "Мобильное приложение для отслеживания личного углеродного следа. Используем ML для анализа повседневных привычек и даём персональные рекомендации по снижению воздействия на экологию. Уже 5000+ пользователей в бета-версии.", "Мобильные приложения", "экология,ML,мобильное приложение"},
		{3, "DesignHub — маркетплейс дизайн-шаблонов", "Платформа где дизайнеры могут продавать и покупать готовые UI-киты, иконки, шаблоны сайтов. Конкурент Dribbble но с фокусом на белорусский рынок.", "Веб-сервисы", "дизайн,маркетплейс,UI"},
		{4, "TaskFlow — Kanban для стартапов", "Простой и бесплатный Kanban-инструмент, заточенный под нужды стартапов. Интеграция с Telegram, автоматические стендапы, трекинг velocity команды.", "Веб-сервисы", "управление,agile,kanban"},
		{5, "InvestMinsk — платформа для ангельских инвестиций", "Связываем белорусские стартапы с ангельскими инвесторами. Верификация проектов, due diligence, юридическая поддержка сделок.", "Финтех", "инвестиции,финтех,стартапы"},
		{6, "NLPBy — NLP для белорусского языка", "Открытая библиотека обработки естественного языка для белорусского языка. Токенизация, POS-теги, NER, sentiment analysis.", "ИИ и ML", "NLP,белорусский язык,ML,open-source"},
		{7, "BrandPulse — мониторинг упоминаний бренда", "SaaS-сервис для мониторинга упоминаний вашего бренда в соцсетях, СМИ, форумах. Анализ тональности, конкурентный анализ, еженедельные отчёты.", "Веб-сервисы", "маркетинг,аналитика,SaaS"},
		{8, "GoMicro — фреймворк микросервисов", "Легковесный Go-фреймворк для создания микросервисов. Встроенный service discovery, circuit breaker, distributed tracing.", "Веб-сервисы", "Go,микросервисы,open-source"},
		{9, "DataLens — визуализация данных для бизнеса", "Дашборды и визуализации для бизнес-аналитики. Подключение к любым источникам данных, конструктор отчётов, AI-инсайты.", "Веб-сервисы", "аналитика,визуализация,BI"},
		{10, "ReactUI.by — библиотека UI-компонентов", "Открытая библиотека React-компонентов в стиле белорусского дизайна. Доступность, тёмная тема, кастомизация через CSS-переменные.", "Веб-сервисы", "React,UI,компоненты,open-source"},
		{11, "HireMinsk — рекрутинг-платформа", "AI-powered рекрутинг для IT-компаний Беларуси. Автоматический подбор кандидатов, видео-интервью, аналитика воронки найма.", "Веб-сервисы", "HR,рекрутинг,AI"},
		{12, "CloudDeploy — CI/CD для стартапов", "Простой CI/CD сервис для небольших команд. Деплой в один клик, интеграция с GitHub, бесплатный тир для стартапов.", "Веб-сервисы", "DevOps,CI/CD,облако"},
		{13, "ContentAI — генерация контента", "AI-помощник для создания маркетингового контента. Посты для соцсетей, email-рассылки, описания продуктов. Поддержка русского и белорусского.", "ИИ и ML", "AI,контент,маркетинг"},
		{14, "SecAudit — автоматический аудит безопасности", "SaaS для автоматического сканирования веб-приложений на уязвимости. OWASP Top 10, отчёты для compliance, интеграция с CI/CD.", "Веб-сервисы", "безопасность,аудит,SaaS"},
		{15, "TestBot — автоматизация тестирования", "No-code платформа для создания автоматизированных тестов. Запись действий в браузере, параллельный запуск, интеграция с Jira.", "Веб-сервисы", "тестирование,автоматизация,no-code"},
		{16, "MinskQuest — AR-квесты по городу", "Мобильное приложение с AR-квестами по достопримечательностям Минска. Геймификация туризма, коллекционные NFT-значки.", "Игры", "AR,игры,туризм"},
		{17, "EduStart — онлайн-курсы для стартаперов", "Образовательная платформа с курсами от успешных белорусских предпринимателей. Менторство, нетворкинг, практические задания.", "Образование", "образование,стартапы,менторство"},
		{18, "MedTech — телемедицина для регионов", "Платформа телемедицины для жителей малых городов. Консультации врачей онлайн, электронные рецепты, интеграция с аптеками.", "Здоровье", "медицина,телемедицина,здоровье"},
		{19, "LogiTrack — оптимизация логистики", "ML-система для оптимизации маршрутов доставки. Экономия до 30% на логистике. Уже используется 15 компаниями в Минске.", "Логистика", "логистика,ML,оптимизация"},
		{20, "FinBot — AI финансовый ассистент", "Чат-бот для управления личными финансами. Категоризация расходов, инвестиционные рекомендации, интеграция с банками Беларуси.", "Финтех", "финтех,AI,чат-бот"},
	}

	for _, p := range posts {
		_, err := db.Exec(
			`INSERT INTO posts (author_id, title, description, category, tags) VALUES ($1, $2, $3, $4, $5)`,
			p.authorID, p.title, p.desc, p.category, p.tags,
		)
		if err != nil {
			return fmt.Errorf("seed post: %w", err)
		}
	}

	comments := []struct {
		postID, authorID int
		content          string
	}{
		{1, 5, "Отличная идея! Экология — это тренд. Какая модель монетизации?"},
		{1, 3, "Дизайн приложения выглядит свежо. Можно посмотреть прототип?"},
		{1, 7, "5000 пользователей за бету — впечатляет. Какой retention?"},
		{2, 10, "Как React-разработчик — жду не дождусь белорусских UI-китов!"},
		{2, 2, "Интересная ниша. Чем отличаетесь от Creative Market?"},
		{3, 8, "Kanban + Telegram — идеальная связка. Когда релиз?"},
		{3, 11, "Нам в HR такой инструмент очень нужен для трекинга задач"},
		{4, 17, "Инвестиционная платформа — это то что нужно экосистеме"},
		{4, 20, "Какой средний размер сделки планируете?"},
		{5, 21, "NLP для белорусского — очень важный проект! Готова помочь с датасетами"},
		{5, 9, "Будет ли API для интеграции в другие сервисы?"},
		{6, 4, "BrandPulse выглядит как белорусский Brandwatch. Какие источники мониторите?"},
		{7, 12, "Как DevOps могу сказать — Go для микросервисов это правильный выбор"},
		{8, 6, "ML для бизнес-аналитики — перспективное направление"},
		{9, 2, "React-компоненты в белорусском стиле — очень интересно!"},
		{10, 4, "AI-рекрутинг сейчас на подъёме. Какая точность подбора?"},
		{11, 8, "CI/CD для стартапов — нужная вещь. Цены конкурентные?"},
		{12, 7, "AI-контент — двоякая тема. Как обеспечиваете качество?"},
		{13, 15, "OWASP сканирование — очень нужно рынку. Какие языки поддерживаете?"},
		{14, 12, "No-code тестирование — это будущее QA"},
		{15, 16, "AR-квесты по Минску — беру! Когда на iOS?"},
		{16, 4, "Менторство от практиков — лучшая форма обучения"},
		{17, 9, "Телемедицина для регионов — социально важный проект"},
		{18, 6, "30% экономии на логистике — впечатляющие цифры. Есть кейсы?"},
		{1, 8, "Присоединяюсь — экология в IT очень актуальна"},
		{2, 6, "Я бы добавил AI-генерацию вариаций дизайна"},
		{3, 2, "Мы используем похожий подход. Могу поделиться опытом"},
		{5, 13, "NLP проект — подписалась! Хочу писать контент на беларускай мове"},
		{7, 15, "Безопасность микросервисов — отдельная тема. Планируете?"},
		{10, 3, "Как дизайнер могу помочь с компонентами. Где контрибьютить?"},
	}

	for _, c := range comments {
		_, err := db.Exec(
			"INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, $3)",
			c.postID, c.authorID, c.content,
		)
		if err != nil {
			return fmt.Errorf("seed comment: %w", err)
		}
	}

	ratings := []struct {
		postID, userID, score int
		review                string
		isExpert              bool
	}{
		{1, 5, 9, "Сильный проект с чёткой миссией. ML-компонент выглядит продуманно. Рекомендую для инвестирования.", true},
		{1, 3, 8, "", false},
		{1, 7, 7, "", false},
		{2, 10, 8, "Хорошая идея, рынок есть", false},
		{2, 5, 7, "Нужна более чёткая бизнес-модель", true},
		{3, 8, 9, "Очень полезный инструмент", false},
		{4, 14, 8, "Платформа нужна экосистеме. Юридическая часть продумана грамотно.", true},
		{4, 17, 9, "", false},
		{5, 9, 10, "Критически важный проект для белорусской культуры", false},
		{5, 14, 9, "Open-source NLP — отлично. Качество моделей на уровне.", true},
		{6, 4, 7, "", false},
		{7, 12, 8, "Хороший выбор технологий", false},
		{8, 6, 8, "", false},
		{9, 2, 9, "Библиотека компонентов — must have", false},
		{10, 11, 8, "", false},
		{11, 8, 7, "", false},
		{14, 12, 9, "Отличное качество сканирования", false},
		{19, 5, 8, "ML оптимизация логистики — технологически сильно. Результаты подтверждены.", true},
	}

	for _, r := range ratings {
		_, err := db.Exec(
			"INSERT INTO ratings (post_id, user_id, score, review, is_expert) VALUES ($1, $2, $3, $4, $5)",
			r.postID, r.userID, r.score, r.review, r.isExpert,
		)
		if err != nil {
			return fmt.Errorf("seed rating: %w", err)
		}
	}

	messages := []struct {
		senderID, receiverID int
		content              string
	}{
		{2, 3, "Привет! Видел твои работы по дизайну. Не хочешь поработать над UI для EcoTrack?"},
		{3, 2, "Привет! Да, интересно. Скинь подробнее ТЗ"},
		{2, 3, "Отлично! Вечером скину макеты и описание"},
		{3, 2, "Жду!"},
		{4, 8, "Сергей, привет! Видел твой GoMicro фреймворк. Крутая работа!"},
		{8, 4, "Спасибо! Если интересно — присоединяйся к контрибьюторам"},
		{5, 17, "Светлана, давайте обсудим инвестиции в ваш проект"},
		{17, 5, "Конечно! Когда удобно созвониться?"},
		{6, 21, "Полина, привет! Работаю над NLP для белорусского. Может объединим усилия?"},
		{21, 6, "Отличная идея! У меня есть размеченные датасеты"},
		{7, 13, "Юлия, можете протестировать BrandPulse на вашем контенте?"},
		{13, 7, "С удовольствием! Пришлите доступ"},
		{10, 16, "Артём! Видел твой AR-проект. Как дела с релизом?"},
		{16, 10, "Работаем над iOS версией. Нужен React Native разработчик"},
		{14, 15, "Наталья, хотите интегрировать SecAudit с TestBot?"},
		{15, 14, "Было бы классно! Давайте обсудим API"},
	}

	for _, m := range messages {
		_, err := db.Exec(
			"INSERT INTO chat_messages (sender_id, receiver_id, content) VALUES ($1, $2, $3)",
			m.senderID, m.receiverID, m.content,
		)
		if err != nil {
			return fmt.Errorf("seed message: %w", err)
		}
	}

	teamRequests := []struct {
		postID, authorID                        int
		title, desc, skills, roleNeeded string
	}{
		{1, 2, "Нужен мобильный разработчик", "Ищем Flutter/React Native разработчика для EcoTrack", "Flutter,React Native,мобильная разработка", "Mobile Developer"},
		{2, 3, "Ищем backend-разработчика", "Нужен Node.js или Go разработчик для маркетплейса", "Node.js,Go,PostgreSQL", "Backend Developer"},
		{3, 4, "Нужен frontend-разработчик", "Ищем React разработчика для TaskFlow", "React,TypeScript,Redux", "Frontend Developer"},
		{5, 6, "Нужен лингвист", "Ищем специалиста по белорусскому языку для разметки данных", "лингвистика,NLP,белорусский язык", "Лингвист"},
		{16, 16, "Нужен 3D-художник", "Ищем 3D-моделлера для AR-объектов", "Blender,Unity,3D", "3D Artist"},
		{18, 19, "Нужен врач-консультант", "Ищем врача для консультирования по медицинским вопросам", "медицина,телемедицина", "Врач-консультант"},
		{8, 8, "Нужен техписатель", "Ищем человека для документации GoMicro", "Go,документация,английский", "Technical Writer"},
		{19, 2, "ML-инженер для логистики", "Нужен ML-инженер для оптимизации алгоритмов", "Python,ML,оптимизация", "ML Engineer"},
	}

	for _, tr := range teamRequests {
		_, err := db.Exec(
			"INSERT INTO team_requests (post_id, author_id, title, description, skills, role_needed) VALUES ($1, $2, $3, $4, $5, $6)",
			tr.postID, tr.authorID, tr.title, tr.desc, tr.skills, tr.roleNeeded,
		)
		if err != nil {
			return fmt.Errorf("seed team request: %w", err)
		}
	}

	teamResponses := []struct {
		requestID, userID int
		message, status   string
	}{
		{1, 19, "Имею опыт с Flutter. Могу подключиться!", "pending"},
		{2, 8, "Могу помочь с Go backend", "accepted"},
		{3, 10, "React — мой основной стек. Заинтересован!", "pending"},
		{4, 21, "Я — AI-исследователь с фокусом на NLP. Готова помочь!", "accepted"},
	}

	for _, tr := range teamResponses {
		_, err := db.Exec(
			"INSERT INTO team_responses (request_id, user_id, message, status) VALUES ($1, $2, $3, $4)",
			tr.requestID, tr.userID, tr.message, tr.status,
		)
		if err != nil {
			return fmt.Errorf("seed team response: %w", err)
		}
	}

	// Follows
	follows := [][2]int{
		{2, 3}, {2, 5}, {2, 8}, {3, 2}, {3, 10}, {4, 2}, {4, 8},
		{5, 2}, {5, 4}, {5, 17}, {6, 5}, {6, 21}, {7, 2}, {7, 13},
		{8, 2}, {8, 4}, {9, 6}, {10, 2}, {10, 3}, {10, 16},
		{11, 4}, {12, 8}, {13, 7}, {14, 15}, {15, 14}, {16, 10},
		{17, 5}, {18, 19}, {19, 2}, {20, 6}, {21, 6}, {21, 5},
	}

	for _, f := range follows {
		db.Exec("INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", f[0], f[1])
	}

	// Friendships (legacy compat)
	friendships := []struct {
		userID, friendID int
		status           string
	}{
		{2, 3, "accepted"}, {3, 2, "accepted"},
		{2, 8, "accepted"}, {8, 2, "accepted"},
		{4, 8, "accepted"}, {8, 4, "accepted"},
		{5, 17, "accepted"}, {17, 5, "accepted"},
		{6, 21, "accepted"}, {21, 6, "accepted"},
		{10, 16, "accepted"}, {16, 10, "accepted"},
		{14, 15, "pending"},
		{7, 2, "pending"},
	}

	for _, f := range friendships {
		db.Exec(
			"INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			f.userID, f.friendID, f.status,
		)
	}

	// Notifications
	notifications := []struct {
		userID                       int
		typ, content, link string
	}{
		{2, "comment", "olga_expert оставил комментарий к вашему посту", "/post/1"},
		{2, "rating", "olga_expert оценил ваш проект на 9", "/post/1"},
		{2, "follow", "maria_design подписалась на вас", "/profile/3"},
		{3, "comment", "nikita_front оставил комментарий к вашему посту", "/post/2"},
		{5, "follow", "dmitry_ml подписался на вас", "/profile/6"},
		{8, "comment", "alex_pm оставил комментарий к вашему посту", "/post/8"},
		{6, "comment", "polina_ai оставил комментарий к вашему посту", "/post/5"},
		{17, "message", "Новое сообщение от olga_expert", "/chat/5"},
		{2, "team_response", "daria_ux откликнулась на ваш запрос команды", "/profile/19"},
		{3, "message", "Новое сообщение от ivan_dev", "/chat/2"},
	}

	for _, n := range notifications {
		db.Exec(
			"INSERT INTO notifications (user_id, type, content, link) VALUES ($1, $2, $3, $4)",
			n.userID, n.typ, n.content, n.link,
		)
	}

	// Complaints
	complaints := []struct {
		authorID, targetID int
		targetType, cat, desc string
	}{
		{11, 12, "post", "спам", "Контент не соответствует тематике платформы"},
		{9, 7, "comment", "оскорбление", "Некорректное высказывание в комментарии"},
		{15, 3, "user", "другое", "Подозрительная активность аккаунта"},
	}

	for _, c := range complaints {
		db.Exec(
			"INSERT INTO complaints (author_id, target_type, target_id, category, description) VALUES ($1, $2, $3, $4, $5)",
			c.authorID, c.targetType, c.targetID, c.cat, c.desc,
		)
	}

	// Expert applications
	expertApps := []struct {
		userID     int
		portfolio, desc, status string
	}{
		{9, "https://portfolio.elena.by", "10 лет опыта в аналитике данных, работала в Яндексе и EPAM", "pending"},
		{12, "https://github.com/pavelfed", "DevOps-инженер с сертификатами AWS и GCP, 7 лет опыта", "pending"},
		{17, "https://light-startup.by", "Основала 2 стартапа с оценкой $1M+, ментор TechMinsk", "pending"},
		{20, "https://roman-arch.dev", "CTO с 12 годами опыта, архитектор систем для 100K+ пользователей", "pending"},
	}

	for _, ea := range expertApps {
		db.Exec(
			"INSERT INTO expert_applications (user_id, portfolio, description, status) VALUES ($1, $2, $3, $4)",
			ea.userID, ea.portfolio, ea.desc, ea.status,
		)
	}

	// Subscriptions
	db.Exec("INSERT INTO subscriptions (user_id, plan, status, expires_at) VALUES (3, 'monthly', 'active', NOW() + INTERVAL '25 days')")
	db.Exec("INSERT INTO subscriptions (user_id, plan, status, expires_at) VALUES (7, 'yearly', 'active', NOW() + INTERVAL '300 days')")
	db.Exec("INSERT INTO subscriptions (user_id, plan, status, expires_at) VALUES (17, 'monthly', 'active', NOW() + INTERVAL '15 days')")

	// Polls
	var pollID1, pollID2 int
	db.QueryRow("INSERT INTO polls (post_id, question) VALUES (1, 'Какая платформа приоритетнее для EcoTrack?') RETURNING id").Scan(&pollID1)
	db.Exec("INSERT INTO poll_options (poll_id, text, vote_count) VALUES ($1, 'iOS', 5)", pollID1)
	db.Exec("INSERT INTO poll_options (poll_id, text, vote_count) VALUES ($1, 'Android', 8)", pollID1)
	db.Exec("INSERT INTO poll_options (poll_id, text, vote_count) VALUES ($1, 'Обе одновременно', 3)", pollID1)

	db.QueryRow("INSERT INTO polls (post_id, question) VALUES (3, 'Какую интеграцию добавить в TaskFlow первой?') RETURNING id").Scan(&pollID2)
	db.Exec("INSERT INTO poll_options (poll_id, text, vote_count) VALUES ($1, 'Slack', 4)", pollID2)
	db.Exec("INSERT INTO poll_options (poll_id, text, vote_count) VALUES ($1, 'Telegram', 12)", pollID2)
	db.Exec("INSERT INTO poll_options (poll_id, text, vote_count) VALUES ($1, 'Discord', 2)", pollID2)

	// Blocked users
	db.Exec("INSERT INTO blocked_users (user_id, blocked_id) VALUES (11, 13) ON CONFLICT DO NOTHING")

	// Projects
	var prjID1, prjID2 int
	db.QueryRow("INSERT INTO projects (user_id, title, description) VALUES (2, 'Экологические приложения', 'Коллекция проектов по экологии') RETURNING id").Scan(&prjID1)
	db.Exec("INSERT INTO project_posts (project_id, post_id) VALUES ($1, 1)", prjID1)

	db.QueryRow("INSERT INTO projects (user_id, title, description) VALUES (5, 'Инвестиционные проекты', 'Проекты рекомендованные к инвестированию') RETURNING id").Scan(&prjID2)
	db.Exec("INSERT INTO project_posts (project_id, post_id) VALUES ($1, 4)", prjID2)
	db.Exec("INSERT INTO project_posts (project_id, post_id) VALUES ($1, 19)", prjID2)

	// Activity logs
	db.Exec("INSERT INTO activity_logs (user_id, action, target_type, target_id, details, ip_address) VALUES (2, 'create_post', 'post', 1, 'EcoTrack', '127.0.0.1')")
	db.Exec("INSERT INTO activity_logs (user_id, action, target_type, target_id, details, ip_address) VALUES (3, 'create_post', 'post', 2, 'DesignHub', '127.0.0.1')")
	db.Exec("INSERT INTO activity_logs (user_id, action, target_type, target_id, details, ip_address) VALUES (5, 'rate_post', 'post', 1, '9', '127.0.0.1')")
	db.Exec("INSERT INTO activity_logs (user_id, action, target_type, target_id, details, ip_address) VALUES (1, 'admin_change_role', 'user', 5, 'expert', '127.0.0.1')")

	log.Println("Database seeded successfully!")
	return nil
}
