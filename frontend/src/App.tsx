import { useState } from "react"

function App() {
  const [region, setRegion] = useState('')

  function handleSearch() {
    const cleanRegion = region.trim()
    if(cleanRegion === ''){
      alert('Вы не ввели регион !')
    }
    else{
      alert(`Вы ищете регион ${cleanRegion}`)
    }
  }

  return(
    <div>
      <h1>Анализатор логистики</h1>
      <p>Система анализа транспортной инфраструктуры</p>

      <input 
       type="text" 
       placeholder="Введите регион"
       value={region}
       onChange={(event)=>setRegion(event.target.value)}
      />

      <button onClick={handleSearch}>
        Найти регион
      </button>
    </div>
  )
}

export default App