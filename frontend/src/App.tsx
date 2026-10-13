function App() {
  function Search() {
    alert('Поиск региона')
  }

  return(
    <div>
      <h1>Анализатор логистики</h1>
      <p>Система анализа транспортной инфраструктуры</p>

      <input type="text" placeholder="Введите регион"/>

      <button onClick={Search}>
        Найти регион
      </button>
    </div>
  )
}

export default App