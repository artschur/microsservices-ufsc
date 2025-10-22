# microsservices-ufsc

seguir arquivo usuarios/main.go

ingressos
  - valida ingresso (recebe userId verifica na tabela de ingressos se userId existe, se existir e tipo de ingresso for especifico, decrementar numero de usos daquele ingresso, retornar 200, senao retornar 400)
  - venda de ingressos (recebe tipo de ingresso, e userId no json, cria um ingresso na tabela de ingressos com o userId e tipo de ingresso)
  - caso o tipo de ingresso for o de numero de usos limitado, botar default de 5.

gateway
  - se comecar com /ingressos encaminha para porta dada no docker compose (exemplo)
  - eh como se fosse apenas um middleware

fila
  - guardar numero de pessoas na fila por atracao, e incrementar/decrementar conforme o tempo de espera da atracao.
  - spawanar goroutine que decrementa um da fila a cada tempo medio de duracao da atracao
  ```
  func main() {
	fmt.Println("Scheduling delayed tasks...")

	// Schedule a task to run after 2 seconds
	time.AfterFunc(2*time.Second, func() {
		fmt.Println("Task 1 executed after 2 seconds")
	})

	// Schedule another task to run after 1 second
	time.AfterFunc(1*time.Second, func() {
		fmt.Println("Task 2 executed after 1 second")
	})

	// Keep the main goroutine alive to allow delayed tasks to execute
	time.Sleep(3 * time.Second)
	fmt.Println("Main function finished.")
  }
  ```
  - recebe json com userId e atracaoId, e incrementa a fila daquela atracao
  - vai mandar post para endpoints de ingresso (valida), para decrementar uso do ingresso do usuario.

espera
  - criar cron(ticker) que a cada 30 segundos verifica o numero de pessoas na fila, quanto tempo a atracao demora para cada pessoa, e salva num hashmap onde chave = atracaoId o tempo de espera daquela atracao (numero de pessoas na fila * tempo medio de duracao da atracao)
  - recebe um get com queryParam atracaoId e retorna o tempo de espera salvo daquela atracao

atracoes
 - tem tempo medio de duracao, nome, id, capacidade
 - GET que retorna todas as atracoes
 - GET que consome da fila o numero de pessoas na fila daquela atracao
usuario
  - criar novo user
