const fs = require('fs');

const data = [
  { name: "Deckers (Hoka)", value: 106.35 },
  { name: "Adidas", value: 80.67 },
  { name: "Nike", value: 44.04 },
  { name: "On Holding", value: 33.81 }
];

const maxValue = 120;
const originX = 120;
const originY = 480;
const chartWidth = 800;
const chartHeight = 350;
const barCount = data.length;
const slotWidth = chartWidth / barCount;
const barGap = slotWidth * 0.3;
const barWidth = slotWidth - barGap;

const nodes = [
  { type: "rect", x: 0, y: 0, width: 1000, height: 600, fillColor: "#F5F6F7", borderColor: "transparent" },
  { type: "text", x: 120, y: 30, width: 800, height: "fit-content", text: "核心竞品最新股价对比 (USD)", fontSize: 24, textAlign: "center" },
  { type: "text", x: 40, y: 80, width: 80, height: "fit-content", text: "股价 (美元)", fontSize: 14, textAlign: "center" },
  
  // Y axis (as rect)
  { type: "rect", x: originX - 1, y: originY - chartHeight - 40, width: 2, height: chartHeight + 40, fillColor: "#333333", borderColor: "transparent" },
  // X axis (as rect)
  { type: "rect", x: originX - 1, y: originY - 1, width: chartWidth + 40, height: 2, fillColor: "#333333", borderColor: "transparent" }
];

// Grid and ticks
const tickCount = 4;
for (let i = 0; i <= tickCount; i++) {
  const tickValue = (maxValue / tickCount) * i;
  const gridY = originY - (tickValue / maxValue) * chartHeight;
  
  // Tick text
  nodes.push({ type: "text", x: originX - 60, y: gridY - 10, width: 50, height: 20, text: String(tickValue), fontSize: 14, textAlign: "right" });
  
  // Tick line (as rect)
  nodes.push({ type: "rect", x: originX - 10, y: gridY - 0.5, width: 10, height: 1, fillColor: "#333333", borderColor: "transparent" });
  
  // Grid line (as rect, simulated dashed by lower opacity or solid thin line)
  if (i > 0) {
    nodes.push({ type: "rect", x: originX, y: gridY - 0.5, width: chartWidth, height: 1, fillColor: "#CCCCCC", borderColor: "transparent" });
  }
}

// Bars
data.forEach((item, i) => {
  const height = (item.value / maxValue) * chartHeight;
  const y = originY - height;
  const x = originX + i * slotWidth + barGap / 2;

  let color = "#A9B0B7"; // Default gray
  if (item.name === "Nike") color = "#3370FF";
  else if (item.name === "Deckers (Hoka)") color = "#F06932";
  else if (item.name === "On Holding") color = "#F06932";
  else color = "#8F959E"; // Adidas

  // Bar
  nodes.push({ type: "rect", id: `bar-${i}`, x: x, y: y, width: barWidth, height: height, fillColor: color, borderColor: "transparent", borderRadius: 4 });
  // Value
  nodes.push({ type: "text", x: x, y: y - 25, width: barWidth, height: 20, text: "$" + item.value.toFixed(2), fontSize: 16, textAlign: "center", textColor: color });
  // Label
  nodes.push({ type: "text", x: x, y: originY + 15, width: barWidth, height: 30, text: item.name, fontSize: 16, textAlign: "center" });
});

const dsl = { version: 2, nodes: nodes };
fs.writeFileSync('bar-chart.json', JSON.stringify(dsl, null, 2));
