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

func normalizeGenre(g string) string {
	switch g {
	case "accounting":
		return "banking"
	case "action":
		return "det_action"
	case "adv_history_avant":
		return "adv_history"
	case "aphorism_quote":
		return "aphorisms"
	case "beginning_authors":
		return ""
	case "biograpy": // misspelling
		return "nonf_biography"
	case "business":
		return "economics_ref"
	case "ci_history": // misspelling
		return "sci_history"
	case "cinema_theatre":
		return "cine" // also theatre
	case "city_fantasy":
		return "sf_fantasy"
	case "comp_programming":
		return "comp_db"
	case "det_cozy":
		return "detective"
	case "dissident":
		return ""
	case "dragon_fantasy":
		return "sf_fantasy"
	case "epic_poetry":
		return "poem"
	case "essay":
		fallthrough
	case "essays":
		return "story"
	case "experimental_poetry":
		return "palindromes"
	case "extravaganza":
		return ""
	case "fable":
		return ""
	case "fanficion": // misspelling
		return "fanfiction"
	case "fantasy":
		return "sf_fantasy"
	case "fantasy_fight":
		return "sf_fantasy"
	case "femslash":
		return "love_erotica"
	case "foreign_action":
		return "det_action"
	case "foreign_adventure":
		return "adventure"
	case "foreign_business":
		return "economics_ref"
	case "foreign_comp":
		return "computers"
	case "foreign_contemporary":
		return "prose_contemporary"
	case "foreign_desc":
		return "reference"
	case "foreign_detective":
		return "detective"
	case "foreign_dramaturgy":
		return "dramaturgy"
	case "foreign_edu":
		return "sci_popular"
	case "foreign_fantasy":
		return "sf_fantasy"
	case "foreign_home":
		return "sci_popular" // not "home"!
	case "foreign_humor":
		return "humor"
	case "foreign_language":
		return "sci_linguistic"
	case "foreign_love":
		return "love"
	case "foreign_novel":
		return "foreign_prose"
	case "foreign_poetry":
		return "poetry"
	case "foreign_psychology":
		return "sci_psychology"
	case "foreign_publicism":
		return "nonf_publicism"
	case "foreign_religion":
		return "religion"
	case "foreign_sf":
		return "sf"
	case "geography_book":
		return "sci_geo"
	case "geo_guide": // misspelling
		return "geo_guides"
	case "global_economy":
		return "sci_economy"
	case "health_rel":
		return ""
	case "historical_fantasy":
		return "sf_fantasy"
	case "humor_fantasy":
		return "sf_fantasy"
	case "industries":
		return ""
	case "in_verse":
		return ""
	case "job_hunting":
		return "popular_business"
	case "literature_rus_classsic": // misspelling
		return "prose_rus_classic"
	case "litrpg":
		return "sf_litrpg"
	case "love_fantasy":
		return "love_sf"
	case "magician_book":
		return "sf_fantasy"
	case "management":
		return "popular_business"
	case "marketing":
		return "org_behavior"
	case "military":
		return "military_weapon" // or military_special, or nonf_military
	case "military_arts":
		return ""
	case "music_dancing":
		return "music"
	case "narrative":
		return "great_story"
	case "newspapers":
		return "periodic"
	case "none": //??
		return ""
	case "nsf": //??
		return ""
	case "palmistry":
		return "astrology"
	case "paper_work":
		return "popular_business"
	case "pedagogy_book":
		return "sci_pedagogy"
	case "personal_finance":
		return "banking"
	case "popadanec":
		return "sf_history"
	case "proce": // misspelling
		return "prose"
	case "prose_epic":
		fallthrough
	case "prose_root":
		return "prose"
	case "prose_rus_classics": // misspelling
		return "prose_rus_classic"
	case "prose_sentimental":
		return "prose"
	case "prose_su_classic": // misspelling
		return "prose_su_classics"
	case "prose_teen":
		return "prose"
	case "psy_alassic": // misspelling (gribuser)
		fallthrough
	case "psy_childs":
		fallthrough
	case "psy_generic":
		fallthrough
	case "psy_personal":
		fallthrough
	case "psy_sex_and_family":
		fallthrough
	case "psy_social":
		fallthrough
	case "psy_theraphy": // misspelling (gribuser)
		return "sci_psychology"
	case "real_estate":
		return ""
	case "rel_boddizm":
		return "religion_budda"
	case "riddles":
		return "folklore"
	case "roman":
		return "great_story"
	case "russian_contemporary":
		return "prose_contemporary"
	case "sagas":
		return "epic"
	case "scenarios":
		return "screenplays"
	case "sci_biochem":
		fallthrough
	case "sci_biophys":
		return "sci_biology"
	case "sci_crib":
		return ""
	case "sci_orgchem":
		fallthrough
	case "sci_physchem":
		return "sci_chem"
	case "sf_all":
		return "sf"
	case "sf_erotic":
		return "love_erotica"
	case "sf_fanfiction":
		return "fanfiction"
	case "sf_fantasy_irony":
		return "sf_fantasy"
	case "sf_history_avant":
		return "sf_history"
	case "sf_irony":
		return "sf_humor"
	case "sf_space_opera":
		return "sf_space"
	case "sf_technofantas": // misspelling
		return "sf_technofantasy"
	case "short_story":
		fallthrough
	case "sketch":
		return "story"
	case "slash":
		return "love_erotica"
	case "small_business":
		return "popular_business"
	case "sociology_book":
		return "sci_social_studies"
	case "stock":
		return "economics"
	case "thriller_legal":
		fallthrough
	case "thriller_medical":
		fallthrough
	case "thriller_techno":
		return "thriller"
	case "unrecognised":
		return ""
	case "upbringing_book":
		return "sci_pedagogy"
	case "vampire_book":
		return "sf"
	case "vers_libre":
		return "palindromes"
	case "visual_arts":
		return "painting"
	case "ya":
		return ""
	}
	return g
}
