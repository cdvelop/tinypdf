package pdf

type ChartFactory struct {
	doc *Document
}

// Chart returns a factory to create various types of charts.
func (d *Document) Chart() *ChartFactory {
	return &ChartFactory{doc: d}
}

// Bar starts building a Bar Chart.
func (f *ChartFactory) Bar() *BarChart {
	return &BarChart{
		doc: f.doc,
	}
}

// Line starts building a Line Chart.
func (f *ChartFactory) Line() *LineChart {
	return &LineChart{
		doc: f.doc,
	}
}

// Pie starts building a Pie Chart.
func (f *ChartFactory) Pie() *PieChart {
	return &PieChart{
		doc: f.doc,
	}
}
