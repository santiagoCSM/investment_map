import { useEffect, useRef } from 'react'

interface MapInteractionState {
  scale: number
  translateX: number
  translateY: number
}

export function useMapInteraction(
  svgRef: React.RefObject<SVGSVGElement>,
  containerWidth: number,
  containerHeight: number,
) {
  const stateRef = useRef<MapInteractionState>({
    scale: 1.3,
    translateX: 0,
    translateY: 0,
  })

  useEffect(() => {
    if (!svgRef.current || containerWidth === 0 || containerHeight === 0) return

    const svg = svgRef.current
    const state = stateRef.current
    const g = svg.querySelector('g') as SVGGElement
    if (!g) return

    const updateTransform = () => {
      g.setAttribute(
        'transform',
        `translate(${state.translateX},${state.translateY}) scale(${state.scale})`,
      )
    }

    const handleWheel = (e: WheelEvent) => {
      if (e.target !== svg && !svg.contains(e.target as Node)) return
      e.preventDefault()

      const rect = svg.getBoundingClientRect()
      const mouseX = e.clientX - rect.left
      const mouseY = e.clientY - rect.top

      const oldScale = state.scale
      const zoomFactor = e.deltaY > 0 ? 0.9 : 1.1
      state.scale = Math.max(1, Math.min(8, state.scale * zoomFactor))

      state.translateX -= (mouseX / oldScale - mouseX / state.scale) * oldScale
      state.translateY -= (mouseY / oldScale - mouseY / state.scale) * oldScale

      updateTransform()
    }

    let isMouseDown = false
    let startX = 0
    let startY = 0

    const handleMouseDown = (e: MouseEvent) => {
      if (e.target !== svg && !svg.contains(e.target as Node)) return
      isMouseDown = true
      startX = e.clientX
      startY = e.clientY
      svg.style.cursor = 'grabbing'
    }

    const handleMouseMove = (e: MouseEvent) => {
      if (!isMouseDown) return
      e.preventDefault()

      const dx = e.clientX - startX
      const dy = e.clientY - startY

      state.translateX += dx
      state.translateY += dy

      startX = e.clientX
      startY = e.clientY

      updateTransform()
    }

    const handleMouseUp = () => {
      isMouseDown = false
      svg.style.cursor = 'grab'
    }

    svg.style.cursor = 'grab'
    updateTransform()

    svg.addEventListener('wheel', handleWheel, { passive: false })
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
    svg.addEventListener('mousedown', handleMouseDown)

    return () => {
      svg.removeEventListener('wheel', handleWheel)
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
      svg.removeEventListener('mousedown', handleMouseDown)
    }
  }, [svgRef, containerWidth, containerHeight])
}
