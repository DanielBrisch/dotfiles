package move

import (
	"testing"

	i3 "go.i3wm.org/i3/v4"
)

func window(id int64, x, y, width, height int64) *i3.Node {
	return &i3.Node{ID: i3.NodeID(id), Type: i3.Con, Rect: i3.Rect{X: x, Y: y, Width: width, Height: height}}
}

func TestNeighbourGradeDeTres(t *testing.T) {
	topLeft := window(1, 0, 0, 500, 500)
	bottomLeft := window(2, 0, 500, 500, 500)
	right := window(3, 500, 0, 500, 1000)
	leaves := []*i3.Node{topLeft, bottomLeft, right}

	cases := []struct {
		name      string
		focused   *i3.Node
		direction string
		want      *i3.Node
	}{
		{"direita a partir do topo esquerdo", topLeft, "right", right},
		{"direita a partir do rodape esquerdo", bottomLeft, "right", right},
		{"esquerda a partir da direita", right, "left", topLeft},
		{"baixo dentro da coluna", topLeft, "down", bottomLeft},
		{"cima dentro da coluna", bottomLeft, "up", topLeft},
		{"sem vizinho a esquerda", topLeft, "left", nil},
		{"sem vizinho acima", topLeft, "up", nil},
		{"sem vizinho abaixo na coluna cheia", right, "down", nil},
	}

	for _, test := range cases {
		got := neighbour(test.focused, leaves, test.direction)
		if got != test.want {
			t.Errorf("%s: peguei %v, quero %v", test.name, nodeID(got), nodeID(test.want))
		}
	}
}

func TestNeighbourEscolheAMaisProxima(t *testing.T) {
	focused := window(1, 0, 0, 300, 1000)
	near := window(2, 300, 0, 300, 1000)
	far := window(3, 600, 0, 400, 1000)

	if got := neighbour(focused, []*i3.Node{focused, far, near}, "right"); got != near {
		t.Errorf("peguei %v, quero a janela mais proxima (%v)", nodeID(got), nodeID(near))
	}
}

func TestNeighbourExigeSobreposicao(t *testing.T) {
	focused := window(1, 0, 0, 500, 400)
	desalinhada := window(2, 500, 400, 500, 400)

	if got := neighbour(focused, []*i3.Node{focused, desalinhada}, "right"); got != nil {
		t.Errorf("peguei %v, quero nil: as janelas nao se sobrepoem no eixo vertical", nodeID(got))
	}
}

func TestRunRejeitaDirecaoInvalida(t *testing.T) {
	if err := Run("diagonal"); err == nil {
		t.Error("quero erro para uma direcao invalida")
	}
}

func nodeID(node *i3.Node) any {
	if node == nil {
		return nil
	}
	return node.ID
}
