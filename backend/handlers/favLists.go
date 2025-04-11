// backend/handlers/favLists.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetLocations returns a list of cities with their coordinates.
func GetLocations(c *gin.Context) {
	locations := []map[string]interface{}{
		{"city": "Helsinki", "latitude": 60.1695, "longitude": 24.9354},
		{"city": "Espoo", "latitude": 60.2055, "longitude": 24.6559},
		{"city": "Tampere", "latitude": 61.4978, "longitude": 23.7610},
		{"city": "Vantaa", "latitude": 60.2934, "longitude": 25.0378},
		{"city": "Oulu", "latitude": 65.0121, "longitude": 25.4651},
		{"city": "Turku", "latitude": 60.4518, "longitude": 22.2666},
		{"city": "Jyväskylä", "latitude": 62.2415, "longitude": 25.7209},
		{"city": "Lahti", "latitude": 60.9827, "longitude": 25.6612},
		{"city": "Kuopio", "latitude": 62.8924, "longitude": 27.6770},
		{"city": "Kouvola", "latitude": 60.8681, "longitude": 26.7042},
		{"city": "Pori", "latitude": 61.4851, "longitude": 21.7976},
		{"city": "Joensuu", "latitude": 62.6012, "longitude": 29.7630},
		{"city": "Lappeenranta", "latitude": 61.0587, "longitude": 28.1887},
		{"city": "Hämeenlinna", "latitude": 61.0078, "longitude": 24.4521},
		{"city": "Vaasa", "latitude": 63.0960, "longitude": 21.6158},
		{"city": "Rovaniemi", "latitude": 66.5039, "longitude": 25.7294},
		{"city": "Seinäjoki", "latitude": 62.7903, "longitude": 22.8405},
		{"city": "Mikkeli", "latitude": 61.6880, "longitude": 27.2722},
		{"city": "Kotka", "latitude": 60.4666, "longitude": 26.9458},
		{"city": "Salo", "latitude": 60.3833, "longitude": 23.1333},
		{"city": "Porvoo", "latitude": 60.3923, "longitude": 25.6653},
		{"city": "Kokkola", "latitude": 63.8385, "longitude": 23.1307},
		{"city": "Hyvinkää", "latitude": 60.6333, "longitude": 24.8667},
		{"city": "Järvenpää", "latitude": 60.4706, "longitude": 25.0899},
		{"city": "Nurmijärvi", "latitude": 60.4641, "longitude": 24.8070},
	}
	c.JSON(http.StatusOK, locations)
}

// GetGenres returns a list of movie genres.
func GetGenres(c *gin.Context) {
	genres := []string{
		"Action", "Adventure", "Animation", "Biography", "Comedy", "Crime", "Documentary",
		"Drama", "Family", "Fantasy", "Film-Noir", "History", "Horror", "Music", "Musical",
		"Mystery", "Romance", "Sci-Fi", "Short", "Sport", "Thriller", "War", "Western",
	}
	c.JSON(http.StatusOK, genres)
}

// GetMovies returns a list of movie titles.
func GetMovies(c *gin.Context) {
	movies := []string{
		"A Beautiful Mind", "A Separation", "All About Eve", "Amadeus", "Annie Hall", "Apollo 13",
		"Ben-Hur", "Birdman", "Braveheart", "Chariots of Fire", "Chicago", "CODA", "Crash",
		"Dances with Wolves", "Departures", "Driving Miss Daisy", "Forrest Gump", "Gladiator",
		"Gone with the Wind", "Green Book", "Hamlet", "In a Better World", "Lawrence of Arabia",
		"Midnight Cowboy", "Moonlight", "Schindler's List", "Slumdog Millionaire", "The Godfather",
		"Parasite", "The Hurt Locker", "The King's Speech", "The Lord of the Rings: The Return of the King",
		"Spotlight", "The Silence of the Lambs", "Titanic", "West Side Story",
	}
	c.JSON(http.StatusOK, movies)
}

// GetDirectors returns a list of movie directors.
func GetDirectors(c *gin.Context) {
	directors := []string{
		"Alfred Hitchcock", "Ang Lee", "Clint Eastwood", "David Lynch", "James Cameron", "Stanley Kubrick",
		"Martin Scorsese", "Quentin Tarantino", "Steven Spielberg", "Ridley Scott", "George Lucas",
		"Woody Allen", "Francis Ford Coppola", "Christopher Nolan", "Wes Anderson", "David Fincher",
	}
	c.JSON(http.StatusOK, directors)
}

// GetActors returns a list of movie actors.
func GetActors(c *gin.Context) {
	actors := []string{
		"Adrien Brody", "Brad Pitt", "Casey Affleck", "Daniel Day-Lewis", "Denzel Washington", "Gary Oldman",
		"Johnny Depp", "Leonardo DiCaprio", "Tom Hanks", "Will Smith", "Robert Downey Jr.", "Russell Crowe",
		"Sean Penn", "Tommy Lee Jones", "Robin Williams", "George Clooney", "James Earl Jones",
	}
	c.JSON(http.StatusOK, actors)
}

// GetActresses returns a list of movie actresses.
func GetActresses(c *gin.Context) {
	actresses := []string{
		"Allison Janney", "Amy Adams", "Angelina Jolie", "Audrey Hepburn", "Bette Davis", "Cate Blanchett",
		"Charlize Theron", "Emily Blunt", "Emma Stone", "Frances McDormand", "Glenn Close", "Halle Berry",
		"Meryl Streep", "Nicole Kidman", "Penélope Cruz", "Reese Witherspoon", "Viola Davis",
	}
	c.JSON(http.StatusOK, actresses)
}
