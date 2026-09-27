package nesting

func orderCrossover(
	parentA Chromosome,
	parentB Chromosome,
	start int,
	end int,
) Chromosome {
	child := Chromosome{
		Genes: make([]Gene, len(parentA.Genes)),
	}

	used := make(map[int]struct{})

	for i := start; i < end; i++ {
		child.Genes[i] = parentA.Genes[i]
		used[child.Genes[i].InstanceID] = struct{}{}
	}

	childIdx := 0

	for _, gene := range parentB.Genes {
		_, exists := used[gene.InstanceID]
		if exists {
			continue
		}

		if childIdx == start {
			childIdx = end
		}

		child.Genes[childIdx] = gene
		childIdx++
	}

	return child
}
