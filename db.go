// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/opennota/fb2index/trigram"
)

var (
	db *sqlx.DB

	ErrNoRows = sql.ErrNoRows
)

func initDB() {
	db = sqlx.MustConnect("sqlite3", *dataSource)
	db.MustExec(`CREATE TABLE IF NOT EXISTS books (
				id              INTEGER PRIMARY KEY AUTOINCREMENT,
				title           TEXT,
				lang            TEXT,
				archive         TEXT,
				filename        TEXT,
				offset          INTEGER,
				compressed_size INTEGER,
				uncompressed_size INTEGER,
				crc32           INTEGER,
				UNIQUE (archive, filename)
			);
			CREATE TABLE IF NOT EXISTS genres (
				id              INTEGER PRIMARY KEY AUTOINCREMENT,
				name            TEXT,
				desc            TEXT,
				meta            TEXT,
				UNIQUE (name)
			);
			CREATE TABLE IF NOT EXISTS authors (
				id              INTEGER PRIMARY KEY AUTOINCREMENT,
				first_name      TEXT,
				middle_name     TEXT,
				last_name       TEXT,
				nickname        TEXT,
				UNIQUE (first_name, middle_name, last_name, nickname)
			);
			CREATE TABLE IF NOT EXISTS sequences (
				id              INTEGER PRIMARY KEY AUTOINCREMENT,
				name            TEXT,
				UNIQUE (name)
			);
			CREATE TABLE IF NOT EXISTS book_genres (
				book_id         INTEGER,
				genre_id        INTEGER,
				PRIMARY KEY (book_id, genre_id)
			);
			CREATE TABLE IF NOT EXISTS book_authors (
				book_id         INTEGER,
				author_id       INTEGER,
				PRIMARY KEY (book_id, author_id)
			);
			CREATE TABLE IF NOT EXISTS book_translators (
				book_id         INTEGER,
				author_id       INTEGER,
				PRIMARY KEY (book_id, author_id)
			);
			CREATE TABLE IF NOT EXISTS book_sequences (
				book_id         INTEGER,
				sequence_id     INTEGER,
				number          INTEGER,
				PRIMARY KEY (book_id, sequence_id)
			);

			CREATE INDEX IF NOT EXISTS books_title_idx ON books (title);
			CREATE INDEX IF NOT EXISTS book_genres_idx ON book_genres (genre_id);
			CREATE INDEX IF NOT EXISTS book_authors_idx ON book_authors (author_id);
			CREATE INDEX IF NOT EXISTS book_translators_idx ON book_translators (author_id);
			CREATE INDEX IF NOT EXISTS book_sequences_idx ON book_sequences (sequence_id);
			CREATE INDEX IF NOT EXISTS authors_idx ON authors (last_name, first_name, nickname);

			INSERT INTO genres (name, desc, meta) VALUES
				('adv_animal', 'Природа и животные', 'Приключения'),
				('adventure', 'Приключения', 'Приключения'),
				('adv_geo', 'Путешествия и география', 'Приключения'),
				('adv_history', 'Исторические приключения', 'Приключения'),
				('adv_indian', 'Вестерн, про индейцев', 'Приключения'),
				('adv_maritime', 'Морские приключения', 'Приключения'),
				('adv_modern', 'Приключения в современном мире', 'Приключения'),
				('adv_story', 'Авантюрный роман', 'Приключения'),
				('antique', 'antique', 'Старинное'),
				('antique_ant', 'Античная литература', 'Старинное'),
				('antique_east', 'Древневосточная литература', 'Старинное'),
				('antique_european', 'Европейская старинная литература', 'Старинное'),
				('antique_myths', 'Мифы. Легенды. Эпос', 'Фольклор'),
				('antique_russian', 'Древнерусская литература', 'Старинное'),
				('aphorisms', 'Афоризмы, цитаты', 'Проза'),
				('architecture_book', 'Скульптура и архитектура', 'Искусство, Искусствоведение, Дизайн'),
				('art_criticism', 'Искусствоведение', 'Искусство, Искусствоведение, Дизайн'),
				('art_world_culture', 'Мировая художественная культура', 'Искусство, Искусствоведение, Дизайн'),
				('astrology', 'Астрология и хиромантия', 'Религия, духовность, эзотерика'),
				('auto_business', 'Автодело', 'Техника'),
				('auto_regulations', 'Автомобили и ПДД', 'Дом и семья'),
				('banking', 'Финансы', 'Деловая литература'),
				('child_adv', 'Приключения для детей и подростков', 'Приключения'),
				('child_classical', 'Классическая детская литература', 'Литература для детей'),
				('child_det', 'Детская остросюжетная литература', 'Литература для детей'),
				('child_education', 'Детская образовательная литература', 'Литература для детей'),
				('child_folklore', 'Детский фольклор', 'Фольклор'),
				('child_prose', 'Проза для детей', 'Литература для детей'),
				('children', 'Детская литература', 'Литература для детей'),
				('child_sf', 'Фантастика для детей', 'Литература для детей'),
				('child_tale_rus', 'Русские сказки', 'Литература для детей'),
				('child_tale', 'Сказки народов мира', 'Литература для детей'),
				('child_verse', 'Стихи для детей', 'Литература для детей'),
				('cine', 'Кино', 'Искусство, Искусствоведение, Дизайн'),
				('comedy', 'Комедия', 'Драматургия'),
				('comics', 'Комиксы', 'Прочее'),
				('comp_db', 'Программирование, программы, базы данных', 'Компьютеры и Интернет'),
				('comp_hard', 'Компьютерное ''железо'' (аппаратное обеспечение), цифровая обработка сигналов', 'Компьютеры и Интернет'),
				('computers', 'Зарубежная компьютерная, околокомпьютерная литература', 'Компьютеры и Интернет'),
				('comp_www', 'ОС и Сети, интернет', 'Компьютеры и Интернет'),
				('design', 'Искусство и Дизайн', 'Искусство, Искусствоведение, Дизайн'),
				('det_action', 'Боевик', 'Детективы и Триллеры'),
				('det_classic', 'Классический детектив', 'Детективы и Триллеры'),
				('det_crime', 'Криминальный детектив', 'Детективы и Триллеры'),
				('detective', 'Детективы', 'Детективы и Триллеры'),
				('det_espionage', 'Шпионский детектив', 'Детективы и Триллеры'),
				('det_hard', 'Крутой детектив', 'Детективы и Триллеры'),
				('det_history', 'Исторический детектив', 'Детективы и Триллеры'),
				('det_irony', 'Иронический детектив, дамский детективный роман', 'Детективы и Триллеры'),
				('det_maniac', 'Про маньяков', 'Детективы и Триллеры'),
				('det_police', 'Полицейский детектив', 'Детективы и Триллеры'),
				('det_political', 'Политический детектив', 'Детективы и Триллеры'),
				('det_su', 'Советский детектив', 'Детективы и Триллеры'),
				('drama_antique', 'Античная драма', 'Драматургия'),
				('dramaturgy', 'Драматургия', 'Драматургия'),
				('drama', 'Драма', 'Драматургия'),
				('economics_ref', 'Деловая литература', 'Деловая литература'),
				('economics', 'Экономика', 'Деловая литература'),
				('epic', 'Былины, эпопея', 'Фольклор'),
				('epistolary_fiction', 'Эпистолярная проза', 'Проза'),
				('equ_history', 'История техники', 'Техника'),
				('fairy_fantasy', 'Мифологическое фэнтези', 'Фантастика'),
				('family', 'Семейные отношения', 'Дом и семья'),
				('fanfiction', 'Фанфик', 'Прочее'),
				('folklore', 'Фольклор, загадки', 'Фольклор'),
				('folk_songs', 'Народные песни', 'Фольклор'),
				('folk_tale', 'Народные сказки', 'Фольклор'),
				('foreign_antique', 'Средневековая классическая проза', 'Проза'),
				('foreign_children', 'Зарубежная литература для детей', 'Литература для детей'),
				('foreign_prose', 'Зарубежная классическая проза', 'Проза'),
				('geo_guides', 'Путеводители, карты, атласы', 'Справочная литература'),
				('gothic_novel', 'Готический роман', 'Проза'),
				('great_story', 'Роман, повесть', 'Проза'),
				('home_collecting', 'Коллекционирование', 'Дом и семья'),
				('home_cooking', 'Кулинария', 'Дом и семья'),
				('home_crafts', 'Хобби и ремесла', 'Дом и семья'),
				('home_diy', 'Сделай сам', 'Дом и семья'),
				('home_entertain', 'Развлечения', 'Дом и семья'),
				('home_garden', 'Сад и огород', 'Дом и семья'),
				('home_health', 'Здоровье', 'Дом и семья'),
				('home_pets', 'Домашние животные', 'Дом и семья'),
				('home_sex', 'Семейные отношения, секс', 'Дом и семья'),
				('home_sport', 'Боевые искусства, спорт', 'Дом и семья'),
				('home', 'Домоводство', 'Дом и семья'),
				('hronoopera', 'Хроноопера', 'Фантастика'),
				('humor_anecdote', 'Анекдоты', 'Юмор'),
				('humor_prose', 'Юмористическая проза', 'Юмор'),
				('humor_satire', 'Сатира', 'Юмор'),
				('humor_verse', 'Юмористические стихи, басни', 'Поэзия'),
				('humor', 'Юмор', 'Юмор'),
				('limerick', 'Частушки, прибаутки, потешки', 'Фольклор'),
				('literature_18', 'Классическая проза XVII-XVIII веков', 'Проза'),
				('literature_19', 'Классическая проза ХIX века', 'Проза'),
				('literature_20', 'Классическая проза ХX века', 'Проза'),
				('love_contemporary', 'Современные любовные романы', 'Любовные романы'),
				('love_detective', 'Остросюжетные любовные романы', 'Любовные романы'),
				('love_erotica', 'Эротическая литература', 'Любовные романы'),
				('love_hard', 'Порно', 'Любовные романы'),
				('love_history', 'Исторические любовные романы', 'Любовные романы'),
				('love_sf', 'Любовное фэнтези, любовно-фантастические романы', 'Любовные романы'),
				('love_short', 'Короткие любовные романы', 'Любовные романы'),
				('love', 'Любовные романы', 'Любовные романы'),
				('lyrics', 'Лирика', 'Поэзия'),
				('military_history', 'Военная история', 'Наука, Образование'),
				('military_special', 'Военное дело', 'Документальная литература'),
				('military_weapon', 'Военное дело, военная техника и вооружение', 'Техника'),
				('modern_tale', 'Современная сказка', 'Фантастика'),
				('music', 'Музыка', 'Искусство, Искусствоведение, Дизайн'),
				('network_literature', 'Самиздат, сетевая литература', 'Прочее'),
				('nonf_biography', 'Биографии и Мемуары', 'Документальная литература'),
				('nonf_criticism', 'Критика', 'Искусство, Искусствоведение, Дизайн'),
				('nonfiction', 'Документальная литература', 'Документальная литература'),
				('nonf_military', 'Военная документалистика и аналитика', 'Документальная литература'),
				('nonf_publicism', 'Публицистика', 'Документальная литература'),
				('notes', 'Партитуры', 'Искусство, Искусствоведение, Дизайн'),
				('org_behavior', 'Маркетинг, PR', 'Деловая литература'),
				('other', 'Неотсортированное', 'Прочее'),
				('painting', 'Живопись, альбомы, иллюстрированные каталоги', 'Искусство, Искусствоведение, Дизайн'),
				('palindromes', 'Визуальная и экспериментальная поэзия, верлибры, палиндромы', 'Поэзия'),
				('periodic', 'Журналы, газеты', 'Прочее'),
				('poem', 'Поэма, эпическая поэзия', 'Поэзия'),
				('poetry_classical', 'Классическая поэзия', 'Поэзия'),
				('poetry_east', 'Поэзия Востока', 'Поэзия'),
				('poetry_for_classical', 'Классическая зарубежная поэзия', 'Поэзия'),
				('poetry_for_modern', 'Современная зарубежная поэзия', 'Поэзия'),
				('poetry_modern', 'Современная поэзия', 'Поэзия'),
				('poetry_rus_classical', 'Классическая русская поэзия', 'Поэзия'),
				('poetry_rus_modern', 'Современная русская поэзия', 'Поэзия'),
				('poetry', 'Поэзия', 'Поэзия'),
				('popular_business', 'Карьера, кадры', 'Деловая литература'),
				('prose_abs', 'Фантасмагория, абсурдистская проза', 'Проза'),
				('prose_classic', 'Классическая проза', 'Проза'),
				('prose_contemporary', 'Современная русская и зарубежная проза', 'Проза'),
				('prose_counter', 'Контркультура', 'Проза'),
				('prose_game', 'Игры, упражнения для детей', 'Литература для детей'),
				('prose_history', 'Историческая проза', 'Проза'),
				('prose_magic', 'Магический реализм', 'Проза'),
				('prose_military', 'Проза о войне', 'Проза'),
				('prose_neformatny', 'Экспериментальная, неформатная проза', 'Проза'),
				('prose_rus_classic', 'Русская классическая проза', 'Проза'),
				('prose_su_classics', 'Советская классическая проза', 'Проза'),
				('prose', 'Проза', 'Проза'),
				('proverbs', 'Пословицы, поговорки', 'Фольклор'),
				('ref_dict', 'Словари', 'Справочная литература'),
				('ref_encyc', 'Энциклопедии', 'Справочная литература'),
				('reference', 'Справочная литература', 'Справочная литература'),
				('ref_guide', 'Руководства', 'Справочная литература'),
				('ref_ref', 'Справочники', 'Справочная литература'),
				('religion_budda', 'Буддизм', 'Религия, духовность, эзотерика'),
				('religion_catholicism', 'Католицизм', 'Религия, духовность, эзотерика'),
				('religion_christianity', 'Христианство', 'Религия, духовность, эзотерика'),
				('religion_esoterics', 'Эзотерика, эзотерическая литература', 'Религия, духовность, эзотерика'),
				('religion_hinduism', 'Индуизм', 'Религия, духовность, эзотерика'),
				('religion_islam', 'Ислам', 'Религия, духовность, эзотерика'),
				('religion_judaism', 'Иудаизм', 'Религия, духовность, эзотерика'),
				('religion_orthodoxy', 'Православие', 'Религия, духовность, эзотерика'),
				('religion_paganism', 'Язычество', 'Религия, духовность, эзотерика'),
				('religion_protestantism', 'Протестантизм', 'Религия, духовность, эзотерика'),
				('religion_self', 'Самосовершенствование', 'Религия, духовность, эзотерика'),
				('religion', 'Религия, религиозная литература', 'Религия, духовность, эзотерика'),
				('russian_fantasy', 'Славянское фэнтези', 'Фантастика'),
				('sci_biology', 'Биология, биофизика, биохимия', 'Наука, Образование'),
				('sci_botany', 'Ботаника', 'Наука, Образование'),
				('sci_build', 'Строительство и сопромат', 'Техника'),
				('sci_chem', 'Химия', 'Наука, Образование'),
				('sci_cosmos', 'Астрономия и Космос', 'Наука, Образование'),
				('sci_culture', 'Культурология', 'Искусство, Искусствоведение, Дизайн'),
				('sci_ecology', 'Экология', 'Наука, Образование'),
				('sci_economy', 'Экономика', 'Наука, Образование'),
				('science', 'Научная литература', 'Наука, Образование'),
				('sci_geo', 'Геология и география', 'Наука, Образование'),
				('sci_history', 'История', 'Наука, Образование'),
				('sci_juris', 'Юриспруденция', 'Наука, Образование'),
				('sci_linguistic', 'Языкознание, иностранные языки', 'Наука, Образование'),
				('sci_math', 'Математика', 'Наука, Образование'),
				('sci_medicine_alternative', 'Альтернативная медицина', 'Наука, Образование'),
				('sci_medicine', 'Медицина', 'Наука, Образование'),
				('sci_metal', 'Металлургия', 'Техника'),
				('sci_oriental', 'Востоковедение', 'Наука, Образование'),
				('sci_pedagogy', 'Педагогика, воспитание детей, литература для родителей', 'Дом и семья'),
				('sci_philology', 'Литературоведение', 'Наука, Образование'),
				('sci_philosophy', 'Философия', 'Наука, Образование'),
				('sci_phys', 'Физика', 'Наука, Образование'),
				('sci_politics', 'Политика', 'Наука, Образование'),
				('sci_popular', 'Зарубежная образовательная литература, зарубежная прикладная,  научно-популярная  литература', 'Наука, Образование'),
				('sci_psychology', 'Психология и психотерапия', 'Наука, Образование'),
				('sci_radio', 'Радиоэлектроника', 'Техника'),
				('sci_religion', 'Религиоведение', 'Религия, духовность, эзотерика'),
				('sci_social_studies', 'Обществознание, социология', 'Наука, Образование'),
				('sci_state', 'Государство и право', 'Наука, Образование'),
				('sci_tech', 'Технические науки', 'Техника'),
				('sci_textbook', 'Учебники и пособия', 'Учебники и пособия'),
				('sci_theories', 'Альтернативные науки и научные теории', 'Наука, Образование'),
				('sci_transport', 'Транспорт и авиация', 'Техника'),
				('sci_veterinary', 'Ветеринария', 'Наука, Образование'),
				('sci_zoo', 'Зоология', 'Наука, Образование'),
				('screenplays', 'Сценарий', 'Драматургия'),
				('sf_action', 'Боевая фантастика', 'Фантастика'),
				('sf_cyberpunk', 'Киберпанк', 'Фантастика'),
				('sf_detective', 'Детективная фантастика', 'Фантастика'),
				('sf_epic', 'Эпическая фантастика', 'Фантастика'),
				('sf_etc', 'Фантастика', 'Фантастика'),
				('sf_fantasy_city', 'Городское фэнтези', 'Фантастика'),
				('sf_fantasy', 'Фэнтези', 'Фантастика'),
				('sf_heroic', 'Героическая фантастика', 'Фантастика'),
				('sf_history', 'Альтернативная история, попаданцы', 'Фантастика'),
				('sf_horror', 'Ужасы', 'Фантастика'),
				('sf_humor', 'Юмористическая фантастика', 'Фантастика'),
				('sf_litrpg', 'ЛитРПГ', 'Фантастика'),
				('sf_mystic', 'Мистика', 'Фантастика'),
				('sf_postapocalyptic', 'Постапокалипсис', 'Фантастика'),
				('sf_social', 'Социально-психологическая фантастика', 'Фантастика'),
				('sf_space', 'Космическая фантастика', 'Фантастика'),
				('sf_stimpank', 'Стимпанк', 'Фантастика'),
				('sf_technofantasy', 'Технофэнтези', 'Фантастика'),
				('sf', 'Научная Фантастика', 'Фантастика'),
				('song_poetry', 'Песенная поэзия', 'Поэзия'),
				('story', 'Малые литературные формы прозы: рассказы, эссе, новеллы, феерия', 'Проза'),
				('tale_chivalry', 'Рыцарский роман', 'Приключения'),
				('tbg_computers', 'Учебные пособия, самоучители', 'Компьютеры и Интернет'),
				('tbg_higher', 'Учебники и пособия ВУЗов', 'Учебники и пособия'),
				('tbg_school', 'Школьные учебники и пособия, рефераты, шпаргалки', 'Учебники и пособия'),
				('tbg_secondary', 'Учебники и пособия для среднего и специального образования', 'Учебники и пособия'),
				('theatre', 'Театр', 'Искусство, Искусствоведение, Дизайн'),
				('thriller', 'Триллер', 'Детективы и Триллеры'),
				('tragedy', 'Трагедия', 'Драматургия'),
				('travel_notes', 'География, путевые заметки', 'Документальная литература'),
				('unfinished', 'Незавершенное', 'Прочее'),
				('vaudeville', 'Мистерия, буффонада, водевиль', 'Драматургия');
	`)
}

func insertWorker(tx *sqlx.Tx, books <-chan book, done chan<- bool) {
	defer close(done)

	for book := range books {
		err := indexBook(tx, book)
		if err != nil {
			log.Printf("%s/%s: failed to add book: %v", book.Archive, book.Filename, err)
		}
	}

	err := tx.Commit()
	if err != nil {
		log.Printf("Commit failed: %v", err)
		tx.Rollback()
	}
}

func startInsertWorker() (chan<- book, <-chan bool) {
	tx := db.MustBegin()

	jobs := make(chan book, *parallel)
	done := make(chan bool)
	go insertWorker(tx, jobs, done)

	return jobs, done
}

func lastInsertID(tx *sqlx.Tx) (id uint32, err error) {
	err = tx.Get(&id, "SELECT last_insert_rowid()")
	return
}

func getOrInsertGenre(tx *sqlx.Tx, g string) (id uint32, inserted bool, err error) {
	err = tx.Get(&id, "SELECT id FROM genres WHERE name = ?", g)
	if err != sql.ErrNoRows {
		return
	}

	_, err = tx.Exec("INSERT INTO genres (name, desc, meta) VALUES (?, '', '')", g)
	if err != nil {
		return
	}

	id, err = lastInsertID(tx)
	return id, err == nil, err
}

func getOrInsertAuthor(tx *sqlx.Tx, a author) (id uint32, inserted bool, err error) {
	err = tx.Get(&id, "SELECT id FROM authors WHERE last_name = ? AND first_name = ? AND middle_name = ? AND nickname = ?",
		a.LastName, a.FirstName, a.MiddleName, a.Nickname)
	if err != sql.ErrNoRows {
		return
	}

	_, err = tx.Exec("INSERT INTO authors (first_name, middle_name, last_name, nickname) VALUES (?, ?, ?, ?)",
		a.FirstName, a.MiddleName, a.LastName, a.Nickname)
	if err != nil {
		return
	}

	id, err = lastInsertID(tx)
	return id, err == nil, err
}

func getOrInsertSequence(tx *sqlx.Tx, name string) (id uint32, inserted bool, err error) {
	err = tx.Get(&id, "SELECT id FROM sequences WHERE name = ?", name)
	if err != sql.ErrNoRows {
		return
	}

	_, err = tx.Exec("INSERT INTO sequences (name) VALUES (?)", name)
	if err != nil {
		return
	}

	id, err = lastInsertID(tx)
	return id, err == nil, err
}

func indexBook(tx *sqlx.Tx, b book) error {
	_, err := tx.Exec("INSERT INTO books (title, lang, archive, filename, offset, compressed_size, uncompressed_size, crc32) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
		b.Title, b.Lang, b.Archive, b.Filename, b.Offset, b.CompressedSize, b.UncompressedSize, b.CRC32)
	if err != nil {
		return err
	}

	bookID, err := lastInsertID(tx)
	if err != nil {
		return err
	}

	var trigrams [][]trigram.T

	for _, g := range b.Genres {
		genreID, _, err := getOrInsertGenre(tx, g.Name)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT OR IGNORE INTO book_genres (book_id, genre_id) VALUES (?, ?)", bookID, genreID)
		if err != nil {
			return err
		}
	}

	for _, a := range b.Authors {
		authorID, inserted, err := getOrInsertAuthor(tx, a)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT OR IGNORE INTO book_authors (book_id, author_id) VALUES (?, ?)", bookID, authorID)
		if err != nil {
			return err
		}

		if inserted {
			for _, s := range []string{a.FirstName, a.MiddleName, a.LastName, a.Nickname} {
				trgm := trigram.Extract(s)
				if len(trgm) > 0 {
					trgmAuthorIndex.AddTrigrams(authorID, trgm)
					trigrams = append(trigrams, trgm)
				}
			}
		}
	}

	for _, a := range b.Translators {
		authorID, inserted, err := getOrInsertAuthor(tx, a)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT OR IGNORE INTO book_translators (book_id, author_id) VALUES (?, ?)", bookID, authorID)
		if err != nil {
			return err
		}

		if inserted {
			for _, s := range []string{a.FirstName, a.MiddleName, a.LastName, a.Nickname} {
				trgm := trigram.Extract(s)
				if len(trgm) > 0 {
					trgmAuthorIndex.AddTrigrams(authorID, trgm)
					trigrams = append(trigrams, trgm)
				}
			}
		}
	}

	for _, s := range b.Sequences {
		seqID, inserted, err := getOrInsertSequence(tx, s.Name)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT OR IGNORE INTO book_sequences (book_id, sequence_id, number) VALUES (?, ?, ?)", bookID, seqID, s.Number)
		if err != nil {
			return err
		}

		if inserted {
			trgm := trigram.Extract(s.Name)
			if len(trgm) > 0 {
				trgmSequenceIndex.AddTrigrams(seqID, trgm)
				trigrams = append(trigrams, trgm)
			}
		}
	}

	for _, trgm := range trigrams {
		trgmBookIndex.AddTrigrams(bookID, trgm)
	}
	trgmBookIndex.Add(bookID, b.Title)

	return nil
}
