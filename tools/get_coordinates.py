import bpy

def find_extreme_points(global_vertices: list):
    """
    Find the most positive and most negative points in a mesh.
    """
    most_positive = [-float('inf'), -float('inf'), -float('inf')]
    most_negative = [float('inf'), float('inf'), float('inf')]

    for v in global_vertices:
        for i in range(3):
            most_positive[i] = max(most_positive[i], v[i])
            most_negative[i] = min(most_negative[i], v[i])
        
    return most_positive, most_negative


def process_meshes(collection):
    """
    Process all meshes in a collection and export their coordinates to a CSV file.
    """
    file_path = "/Users/pierreyves/Programming/Louvain-li-Nux/lln-gamejam-2024/tools/coordinates.csv" # Change this path to your own
    with open(file_path, "w") as file:
        file.write("Name, most_positive_x, most_positive_y, most_positive_z, most_negative_x, most_negative_y, most_negative_z\n")

        for obj in collection.objects:
            mesh = obj.data
            global_vertices = [obj.matrix_world @ v.co for v in mesh.vertices]
            most_positive, most_negative = find_extreme_points(global_vertices)
            file.write(f"{obj.name}, {most_positive[0]}, {most_positive[1]}, {most_positive[2]}, {most_negative[0]}, {most_negative[1]}, {most_negative[2]}\n")

    print(f"Coordinates exported to {file_path}.")

collection = bpy.data.collections.get("colliders")
if collection:
    process_meshes(collection)
else:
    print("Collection not found.")
