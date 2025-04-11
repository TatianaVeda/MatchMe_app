package utils

import (
	"backend/database"
	"backend/models"
	"fmt"
	"math/rand"
	"time"

	"github.com/bxcodec/faker/v3"
)

var genres = []string{
	"Action", "Adventure", "Animation", "Biography", "Comedy", "Crime", "Documentary", "Drama", 
	"Fantasy", "Horror", "Mystery", "Romance", "Sci-Fi", "Thriller", "War", "Western",
}

var movies = []string{
	"Titanic", "The Godfather", "Forrest Gump", "The Dark Knight", "Inception", "Pulp Fiction", 
	"Schindler's List", "The Matrix", "Gladiator", "The Silence of the Lambs",
}

var directors = []string{
	"Christopher Nolan", "Steven Spielberg", "Martin Scorsese", "Quentin Tarantino", 
	"James Cameron", "Alfred Hitchcock", "Stanley Kubrick", "Francis Ford Coppola", "Ridley Scott",
}

var actors = []string{
	"Leonardo DiCaprio", "Brad Pitt", "Johnny Depp", "Robert Downey Jr.", "Tom Hanks", 
	"Marlon Brando", "Al Pacino", "Robert De Niro", "Jack Nicholson", "Denzel Washington",
}

var actresses = []string{
	"Meryl Streep", "Natalie Portman", "Scarlett Johansson", "Jennifer Lawrence", "Emma Stone", 
	"Cate Blanchett", "Julia Roberts", "Sandra Bullock", "Reese Witherspoon",
}

var locationCoordinates = map[string]struct {
	Latitude  float64
	Longitude float64
}{
	"Helsinki": {60.1695, 24.9354}, "Espoo": {60.2055, 24.6559}, "Tampere": {61.4978, 23.7610},
	"Vantaa": {60.2934, 25.0378}, "Oulu": {65.0121, 25.4651}, "Turku": {60.4518, 22.2666},
	"Jyväskylä": {62.2415, 25.7209}, "Lahti": {60.9827, 25.6612}, "Kuopio": {62.8924, 27.6770},
	"Kouvola": {60.8681, 26.7042}, "Pori": {61.4851, 21.7976}, "Joensuu": {62.6012, 29.7630},
	"Lappeenranta": {61.0587, 28.1887}, "Hämeenlinna": {61.0078, 24.4521}, "Vaasa": {63.0960, 21.6158},
	"Rovaniemi": {66.5039, 25.7294}, "Seinäjoki": {62.7903, 22.8405}, "Mikkeli": {61.6880, 27.2722},
	"Kotka": {60.4666, 26.9458}, "Salo": {60.3833, 23.1333}, "Porvoo": {60.3923, 25.6653},
	"Kokkola": {63.8385, 23.1307}, "Hyvinkää": {60.6333, 24.8667}, "Järvenpää": {60.4706, 25.0899},
	"Nurmijärvi": {60.4641, 24.8070},
}

var locations []string
var titles = []string{
	"King", "Queen", "Lord", "Lady", "Duke", "Duchess", "Emperor", "Empress", "Prince", "Princess", 
	"Baron", "Baroness", "Viscount", "Marquess", "Count", "Countess", "Grand Duke", "Grand Duchess",
	"Archduke", "Archduchess", "General", "Colonel", "Major", "Captain", "Commander", "Admiral", 
	"Sergeant", "Lieutenant", "Warrior", "Knight", "Paladin", "Gladiator", "Ranger", "Sentinel", 
	"Guardian", "Spartan", "Samurai", "Ninja", "Viking", "Berserker", "Dragon Slayer", "Demon Hunter", 
	"Archmage", "Necromancer", "Warlock", "Sorcerer", "Summoner", "Druid", "Alchemist", "Shapeshifter", 
	"Shadowmaster", "Spellbinder", "Enchanter", "Beastmaster", "Elementalist", "Time Traveler", "Phoenix King", 
	"Celestial Knight", "Cosmic Lord", "Dark Overlord", "Jedi Master", "Sith Lord", "Bounty Hunter", 
	"Space Cowboy", "Cyberpunk", "Time Lord", "Pirate King", "Gunslinger", "Mad Scientist", "Steampunk Baron", 
	"Captain Chaos", "The Chosen One", "The Shadow", "The Unstoppable", "The Annihilator", "The Invincible", 
	"The Immortal", "The Silent One", "The Phantom", "The Maverick", "Champion", "Grandmaster", 
	"Legendary Hero", "The Titan", "The Gladiator", "The Avenger", "The Protector", "The Crusader", "The Savior", 
	"The Conqueror", "The Juggernaut", "The Guardian Angel", "The Divine Warrior", "The Lone Wolf", 
	"The Stormbringer", "The Firestarter", "The Icebreaker", "The Thunderlord", "The Beast", "The Ultimate Challenger",
}

var usedTitles = make(map[string]bool)

func init() {
	rand.Seed(time.Now().UnixNano())
	shuffleTitles()
	for city := range locationCoordinates {
		locations = append(locations, city)
	}
}

func shuffleTitles() {
	rand.Shuffle(len(titles), func(i, j int) {
		titles[i], titles[j] = titles[j], titles[i]
	})
}

func isUsernameTaken(username string) bool {
	var count int
	row := database.DB.Raw("SELECT COUNT(*) FROM users WHERE username = ?", username).Row()
	row.Scan(&count)
	return count > 0
}

func generateUsername() string {
	if len(usedTitles) >= len(titles) {
		usedTitles = make(map[string]bool)
		shuffleTitles()
	}

	for _, title := range titles {
		username := fmt.Sprintf("%s Chuck Norris", title)
		if !usedTitles[username] && !isUsernameTaken(username) {
			usedTitles[username] = true
			return username
		}
	}
	return fmt.Sprintf("Ultimate Chuck Norris %d", rand.Intn(1000))
}

func getRandomItem(list []string) string {
	return list[rand.Intn(len(list))]
}

func GenerateUsers() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		randomLocation := getRandomItem(locations)
		coordinates := locationCoordinates[randomLocation]
		username := generateUsername()
		user := models.User{
			Username:         username,
			FirstName:        faker.FirstName(),
			LastName:         faker.LastName(),
			Email:            faker.Email(),
			Password:         "password123",
			Location:         randomLocation,
			Latitude:         coordinates.Latitude,
			Longitude:        coordinates.Longitude,
			AboutMe:          faker.Sentence(),
			FavoriteGenre:    getRandomItem(genres),
			FavoriteMovie:    getRandomItem(movies),
			FavoriteDirector: getRandomItem(directors),
			FavoriteActor:    getRandomItem(actors),
			FavoriteActress:  getRandomItem(actresses),
		}
		if err := user.HashPassword(); err == nil {
			database.DB.Create(&user)
		}
	}
}
