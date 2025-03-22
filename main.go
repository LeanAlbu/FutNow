package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Fixture struct {
	Fixture struct {
		ID        int       `json:"id"`
		Date      time.Time `json:"date"`
		Timestamp int       `json:"timestamp"`
		Status    struct {
			Long  string `json:"long"`
			Short string `json:"short"`
		} `json:"status"`
	} `json:"fixture"`
	League struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"league"`
	Teams struct {
		Home Team `json:"home"`
		Away Team `json:"away"`
	} `json:"teams"`
	Goals struct {
		Home *int `json:"home"`
		Away *int `json:"away"`
	} `json:"goals"`
}

type APIResponse struct {
	Results  int       `json:"results"`
	Fixtures []Fixture `json:"response"`
}

type LeagueInfo struct {
	League struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"league"`
}

type LeaguesResponse struct {
	Results int          `json:"results"`
	Leagues []LeagueInfo `json:"response"`
}

func main() {
	apiKey := "b6ed6a7250bf19f5a39f292dc859ad46"

	date := time.Now().Format("2006-01-02")

	// Lista de ligas/competições
	leagues := []struct {
		ID     int
		Season int
		Name   string
	}{
		{ID: 71, Season: time.Now().Year()}, // Serie A (Brasil)
		{ID: 72, Season: time.Now().Year()}, // Serie B (Brasil)
		{ID: 13, Season: time.Now().Year()}, // Copa Libertadores
	}

	for i, league := range leagues {
		leagueName, err := getLeagueNameFromAPI(apiKey, league.ID)
		if err != nil {
			fmt.Printf("Erro ao obter o nome da liga %d: %v\n", league.ID, err)
			leagueName = "Desconhecida"
		}
		leagues[i].Name = leagueName
	}

	for _, league := range leagues {
		url := fmt.Sprintf("https://v3.football.api-sports.io/fixtures?date=%s&league=%d&season=%d", date, league.ID, league.Season)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Erro ao criar requisição para liga %d: %v\n", league.ID, err)
			continue
		}

		// Adicionar o header com a API Key
		req.Header.Add("x-apisports-key", apiKey)

		// Criar um cliente HTTP
		client := &http.Client{}

		// Fazer a requisição
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Erro ao fazer requisição para liga %d: %v\n", league.ID, err)
			resp.Body.Close()
			continue
		}

		// Ler a resposta
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Erro ao ler requisição para liga %d: %v\n", league.ID, err)
			continue
		}

		var apiResponse APIResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			fmt.Printf("Erro ao fazer o parsing do JSON para liga %d: %v\n", league.ID, err)
			continue
		}

		if apiResponse.Results == 0 {
			fmt.Printf("Nenhum jogo encontrado para a liga %s (ID: %d) no dia %s.\n", league.Name, league.ID, date)
			continue
		}

		fmt.Printf("\nJogos do dia %s - Liga: %s (ID: %d)\n", date, league.Name, league.ID)
		fmt.Println("----------------------------------------------------")
		for _, fixture := range apiResponse.Fixtures {
			// Converter o horário para o fuso horário local
			localTime := fixture.Fixture.Date.Local()

			// Exibir informações do jogo
			fmt.Printf("%s x %s\n", fixture.Teams.Home.Name, fixture.Teams.Away.Name)
			fmt.Printf("Horário: %s\n", localTime.Format("15:04"))
			fmt.Printf("Status: %s\n", fixture.Fixture.Status.Long)

			// Exibir o placar, se disponível
			if fixture.Goals.Home != nil && fixture.Goals.Away != nil {
				fmt.Printf("Placar: %d - %d\n", *fixture.Goals.Home, *fixture.Goals.Away)
			} else {
				fmt.Println("Placar: NÃO DISPONÍVEL")
			}
			fmt.Println("----------------------------------------------------")
		}
	}
}

func getLeagueNameFromAPI(apiKey string, leagueID int) (string, error) {
	url := fmt.Sprintf("https://v3.football.api-sports.io/leagues?id=%d", leagueID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Add("x-apisports-key", apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %v", err)
	}

	var leaguesResponse LeaguesResponse
	err = json.Unmarshal(body, &leaguesResponse)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer o parsing do JSON: %v", err)
	}

	if leaguesResponse.Results == 0 || len(leaguesResponse.Leagues) == 0 {
		return "Desconhecida", nil
	}

	return leaguesResponse.Leagues[0].League.Name, nil
}
